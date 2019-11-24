package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vincoll/vigie/pkg/probe"
	"github.com/vincoll/vigie/pkg/utils"
	"golang.org/x/net/http2"
	"net"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"

	"io/ioutil"

	"net/http"
	"os"

	"time"
)

func (p *Probe) process(timeout time.Duration) (probeAnswers []*ProbeAnswer) {

	// Resolve only some IPv
	ips, err := probe.ADVGetIPsfromHostname(p.host, p.IpVersion)
	if err != nil {
		pi := probe.ProbeInfo{Status: probe.Error, Error: err.Error()}
		probeAnswers = make([]*ProbeAnswer, 0, 0)
		probeAnswers = append(probeAnswers, &ProbeAnswer{ProbeInfo: pi})
		return probeAnswers
	}

	if len(ips) == 0 {
		errNoIP := fmt.Errorf("No IP for %s with ipv%d found.", p.host, p.IpVersion)

		pi := probe.ProbeInfo{Status: probe.Error, Error: errNoIP.Error()}
		probeAnswers = make([]*ProbeAnswer, 0, 0)
		probeAnswers = append(probeAnswers, &ProbeAnswer{ProbeInfo: pi})
		return probeAnswers
	}

	// Loop for each ip behind a DNS record
	// probeAnswers store the results for each IP
	probeAnswers = make([]*ProbeAnswer, len(ips))
	var wg sync.WaitGroup
	wg.Add(len(ips))

	for i, ip := range ips {

		go func(i int, ip string) {
			pa, errReq := p.sendTheRequest(ip, timeout)
			if errReq != nil {
				//print(errReq)
			}
			probeAnswers[i] = &pa
			wg.Done()
		}(i, ip)

	}
	wg.Wait()
	return probeAnswers
}

// generateHTTPRequest returns a Go http.request based on all options
func (p Probe) generateHTTPRequest(completeURL string) (*http.Request, error) {

	body := &bytes.Buffer{}

	switch {

	case p.Body != "":
		body = bytes.NewBuffer([]byte(p.Body))

	case p.BodyFile != "":
		path := p.BodyFile
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			temp, err := ioutil.ReadFile(path)
			if err != nil {
				return nil, err
			}
			body = bytes.NewBuffer(temp)
		}
	}

	// Create HTTP New Request
	req, err := http.NewRequest(p.Method, completeURL, body)

	if err != nil {
		return nil, err
	}
	// Add BasicAuth
	if p.BasicAuthUser != "" {
		req.SetBasicAuth(p.BasicAuthUser, p.BasicAuthPassword)
	}

	// Add Headers
	for k, v := range p.Headers {
		req.Header.Set(k, v)
	}

	return req, err
}

func (p Probe) sendTheRequest(ip string, timeout time.Duration) (ProbeAnswer, error) {

	transport, err := p.generateTransport(p.request, ip, timeout)
	if err != nil {
		pi := probe.ProbeInfo{Status: probe.Error, Error: err.Error()}
		pa := ProbeAnswer{ProbeInfo: pi}
		return pa, err
	}

	// Prepare Time Measurements

	var t0DNSStart, t1DNSDone, t2CoDone, t3GotCon, t4FirstByte, t5TLSStart, t6TLSDone time.Time
	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { t0DNSStart = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { t1DNSDone = time.Now() },
		ConnectStart: func(_, _ string) {
			if t1DNSDone.IsZero() {
				// connecting to IP
				t1DNSDone = time.Now()
			}
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				println("unable to connect to host %v: %v", addr, err)
			}
			t2CoDone = time.Now()

		},
		GotConn:              func(_ httptrace.GotConnInfo) { t3GotCon = time.Now() },
		GotFirstResponseByte: func() { t4FirstByte = time.Now() },
		TLSHandshakeStart:    func() { t5TLSStart = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { t6TLSDone = time.Now() },
	}

	request := p.request.WithContext(httptrace.WithClientTrace(context.Background(), trace))

	// Set Client
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	if p.FollowRedirects == false {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			// always refuse to follow redirects, visit does that manually if required.
			return http.ErrUseLastResponse
		}
	}

	// QUICK FIX to add body
	// https://stackoverflow.com/questions/31337891/net-http-http-contentlength-222-with-body-length-0
	request.Body = ioutil.NopCloser(strings.NewReader(p.Body)) //bytes.NewBuffer([]byte(p.Body))

	// SEND REQUEST
	resp, err := client.Do(request)
	t7End := time.Now() // after read body

	if t0DNSStart.IsZero() {
		// we skipped DNS
		t0DNSStart = t1DNSDone
	}

	rst := genResponsesTime(request.URL.Scheme, t0DNSStart, t1DNSDone, t2CoDone, t3GotCon, t4FirstByte, t5TLSStart, t6TLSDone, t7End)
	// Error
	if err != nil {

		pi := probe.ProbeInfo{Status: probe.Error, ResponseTime: t7End.Sub(t0DNSStart), Error: err.Error()}
		pa := ProbeAnswer{ProbeInfo: pi, ResponsesTime: rst}

		return pa, err
	}

	// Success
	pi := probe.ProbeInfo{Status: probe.Success, ResponseTime: t7End.Sub(t0DNSStart)}
	pa := ProbeAnswer{HTTPcode: resp.StatusCode, ProbeInfo: pi, ResponsesTime: rst}

	if resp.Body != nil {
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		generatedName := fmt.Sprintf("%s_(%s)%s", p.GetName(), p.Method, p.URL)
		hashFilename := utils.GetSHA1Hash(generatedName)
		err = saveResponseBody(body, hashFilename)
		if err != nil {
			utils.Log.WithFields(logrus.Fields{
				"package": "probe http",
			}).Errorf("Can't write the response on disk : %v", err)
		}

		if iscontentTypeJSON(resp) {
			bodyJSONArray := []interface{}{}

			if err := json.Unmarshal(body, &bodyJSONArray); err != nil {
				bodyJSONMap := map[string]interface{}{}
				if err2 := json.Unmarshal(body, &bodyJSONMap); err2 == nil {
					pa.BodyJSON = bodyJSONMap
				}
			} else {
				pa.BodyJSON = bodyJSONArray
			}
		} else {
			pa.Body = string(body)
		}

	}

	// Add Headers
	pa.Headers = make(map[string]string, len(resp.Header))
	for k, v := range resp.Header {
		pa.Headers[k] = v[0]
	}

	return pa, nil
}

func keepLines(s string, n int) string {
	result := strings.Join(strings.Split(s, "\n")[:n], "\n")
	return strings.Replace(result, "\r", "", -1)
}

func (p Probe) generateTransport(request *http.Request, ip string, timeout time.Duration) (*http.Transport, error) {

	// Give transport a IP to overwrite DNS resolution

	port := p.request.URL.Port()
	if port == "" {

		switch request.URL.Scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		}
	}

	ipPort := fmt.Sprintf("%s:%s", ip, port)

	tr := &http.Transport{
		// SET PROXY
		//Proxy:                 http.ProxyFromEnvironment,
		IdleConnTimeout:       timeout,
		TLSHandshakeTimeout:   timeout,
		ExpectContinueTimeout: timeout,
	}

	switch p.IpVersion {
	case 4:
		tr.DialContext = dialContext("tcp4", ipPort)
	case 6:
		tr.DialContext = dialContext("tcp6", ipPort)
	default:
		tr.DialContext = dialContext("tcp", ipPort)
	}

	// Add TLS in transport if needed
	switch request.URL.Scheme {
	case "https":
		host, _, err := net.SplitHostPort(request.Host)
		if err != nil {
			// If p.request.Host does not have port, simply add the bare host.
			host = request.Host
		}

		tr.TLSClientConfig = &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: p.IgnoreVerifySSL,
			//Certificates: readClientCert(clientCertFile),
		}

		// Because we create a custom TLSClientConfig, we have to opt-in to HTTP/2.
		// See https://github.com/golang/go/issues/14275
		err = http2.ConfigureTransport(tr)
		if err != nil {
			err = fmt.Errorf("failed to prepare transport for HTTP/2: %v", err)
			return nil, err
		}
	}

	return tr, nil
}

// saveResponseBody consumes the body of the response.
// saveResponseBody returns an informational message about the
// disposition of the response body's contents.
func saveResponseBody(body []byte, filename string) error {

	path := fmt.Sprintf("%s/%s", "/tmp/vigie/tmp", filename)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("unable to create file %s: %v", path, err)
	}
	defer f.Close()

	err = ioutil.WriteFile(path, body, 0660)

	return err
}

func isRedirect(resp *http.Response) bool {
	return resp.StatusCode > 299 && resp.StatusCode < 400
}

func iscontentTypeJSON(resp *http.Response) bool {

	contentType := resp.Header.Get("Content-type")
	if contentType == "application/json" {
		return true
	} else {
		return false
	}
}

func genResponsesTime(scheme string, t0DNSStart, t1DNSDone, t2CoDone, t3GotCon, t4FirstByte, t5TLSStart, t6TLSDone, t7End time.Time) responsestime {

	if scheme == "https" {
		return responsestime{
			dnsLookup:        t1DNSDone.Sub(t0DNSStart),
			tcpConnection:    t2CoDone.Sub(t1DNSDone),
			tlsHandshake:     t6TLSDone.Sub(t5TLSStart),
			serverProcessing: t4FirstByte.Sub(t3GotCon),
			contentTransfert: t7End.Sub(t4FirstByte),
			namelookup:       t1DNSDone.Sub(t0DNSStart),
			connect:          t2CoDone.Sub(t0DNSStart),
			pretransfert:     t3GotCon.Sub(t0DNSStart),
			starttransfert:   t4FirstByte.Sub(t0DNSStart),
			total:            t7End.Sub(t0DNSStart),
		}
	} else {
		// http
		return responsestime{
			dnsLookup:        t1DNSDone.Sub(t0DNSStart),
			tcpConnection:    t3GotCon.Sub(t1DNSDone),
			serverProcessing: t4FirstByte.Sub(t3GotCon),
			contentTransfert: t7End.Sub(t4FirstByte),
			namelookup:       t1DNSDone.Sub(t0DNSStart),
			connect:          t3GotCon.Sub(t0DNSStart),
			starttransfert:   t4FirstByte.Sub(t0DNSStart),
			total:            t7End.Sub(t0DNSStart),
		}
	}

}

func parseURL(uri string) (*url.URL, error) {
	if !strings.Contains(uri, "://") && !strings.HasPrefix(uri, "//") {
		uri = "//" + uri
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		u.Scheme = "http"
		if !strings.HasSuffix(u.Host, ":80") {
			u.Scheme += "s"
		}
	}
	return u, nil
}

// readClientCert - helper function to read client certificate
// from pem formatted file
func readClientCert(filename string) ([]tls.Certificate, error) {
	if filename == "" {
		return nil, nil
	}
	var (
		pkeyPem []byte
		certPem []byte
	)

	// read client certificate file (must include client private key and certificate)
	certFileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read client certificate file: %v", err)
	}

	for {
		block, rest := pem.Decode(certFileBytes)
		if block == nil {
			break
		}
		certFileBytes = rest

		if strings.HasSuffix(block.Type, "PRIVATE KEY") {
			pkeyPem = pem.EncodeToMemory(block)
		}
		if strings.HasSuffix(block.Type, "CERTIFICATE") {
			certPem = pem.EncodeToMemory(block)
		}
	}

	cert, err := tls.X509KeyPair(certPem, pkeyPem)
	if err != nil {
		return nil, fmt.Errorf("unable to load client cert and key pair: %v", err)
	}
	return []tls.Certificate{cert}, nil
}

func dialContext(network, host string) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, _, addr string) (net.Conn, error) {
		a := host
		return (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext(ctx, network, a)
	}
}
