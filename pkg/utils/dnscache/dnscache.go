package dnscache

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Resolver struct {
	// timeout defines the maximum allowed time allowed for a lookup.
	timeout time.Duration

	cache map[string]*cacheEntry

	minCacheDuration time.Duration
	maxCacheDuration time.Duration

	mutex sync.RWMutex

	chanLock        sync.Mutex
	resolveChannels map[string]chan error
}

type cacheEntry struct {
	ips []net.IP
	//	ipsv6 []net.IP For V2
	//	ipsv4 []net.IP

	created        time.Time
	expirationTime time.Time
}

// ip returns an ip and a bool value which indicates whether the cache is valid.
func (i *cacheEntry) ip() ([]net.IP, bool) {
	if len(i.ips) <= 0 {
		return nil, false
	}
	isRecordExpired := time.Now().Before(i.expirationTime)
	return i.ips, isRecordExpired
}

func NewCachedResolver() (*Resolver, error) {

	r := Resolver{
		timeout:          6 * time.Second,
		minCacheDuration: time.Minute,
		maxCacheDuration: 5 * time.Minute,
	}
	r.cache = make(map[string]*cacheEntry, 50)
	r.resolveChannels = make(map[string]chan error)

	return &r, nil
}

// LookupHost looks up the given host using the local resolver. It returns a
// slice of that host's addresses.
func (r *Resolver) LookupHost(ctx context.Context, host string, ipv int) (addrs []string, err error) {

	a, err := r.resolveHost(ctx, host)
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0, len(a))

	switch ipv {

	case 0:
		for _, ip := range a {
			ips = append(ips, ip.String())
		}
	case 4:
		for _, ip := range a {
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	case 6:
		for _, ip := range a {
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}

	}

	return ips, nil
}

func (r *Resolver) GetIPFromHostname(host string, ipv int) (addrs []string, err error) {

	ctx, _ := context.WithTimeout(context.Background(), r.timeout)

	a, err := r.resolveHost(ctx, host)
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0, len(a))

	switch ipv {

	case 0:
		for _, ip := range a {
			ips = append(ips, ip.String())
		}
	case 4:
		for _, ip := range a {
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	case 6:
		for _, ip := range a {
			if ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}

	}

	return ips, nil
}

func (r *Resolver) resolveHost(ctx context.Context, host string) ([]net.IP, error) {
	ips, ok := r.getIPFromCache(host)
	if ok {
		return ips, nil
	}

	r.chanLock.Lock()
	// Concurrent calls : This host peut être entrain de se faire resoudre
	ch := r.resolveChannels[host]
	if ch == nil {
		// Recheck the cache.
		ips, ok := r.getIPFromCache(host)
		if ok {
			return ips, nil
		}
		ch = make(chan error, 1)
		r.resolveChannels[host] = ch
		// There is no resolving process for the host. Create a new goroutine to lookup dns.
		go func() {
			//atomic.AddInt64(&r.stats.DNSQuery, 1)
			var dnsError error
			var item *cacheEntry
			defer func() {
				if item != nil {
					//atomic.AddInt64(&r.stats.SuccessfulDNSQuery, 1)
				}
				r.mutex.Lock()
				if item == nil {
					// Remove host from cache.
					delete(r.cache, host)
				} else {
					// Cache the item.
					r.cache[host] = item
				}
				r.mutex.Unlock()

				// Wake up a resolveHost function.
				r.chanLock.Lock()
				delete(r.resolveChannels, host)
				ch <- dnsError
				r.chanLock.Unlock()
			}()
			// Resolve host without caching
			ipsWOCache, err := net.LookupIP(host)
			if err != nil {
				dnsError = fmt.Errorf("can't resolve host %s because %s", host, err.Error())
				return
			}
			if len(ipsWOCache) <= 0 {
				dnsError = fmt.Errorf("no dns records for host %s", host)
				return
			}
			item = newCacheEntry(ipsWOCache, r.randomEXP())
			return
		}()
	}
	r.chanLock.Unlock()

	select {
	case err := <-ch:
		// Put back the error. This operation can wake up other resolveHost functions.
		ch <- err
		if err != nil {
			return nil, err
		}
		ips, _ := r.getIPFromCache(host)
		// In this case, the dns result is fresh and we can ignore the second result safely.
		if ips != nil {
			return ips, nil
		}
		return nil, fmt.Errorf("no dns records of host %s in cache", host)
	case <-ctx.Done():
		// glog.V(2).Infof("Can't resolve host %s because context timeout", host)
		return nil, ctx.Err()
	}
}

func (r *Resolver) getIPFromCache(host string) ([]net.IP, bool) {
	r.mutex.RLock()
	item, ok := r.cache[host]
	r.mutex.RUnlock()
	if ok {
		return item.ip()
	}
	return nil, false
}

func (r *Resolver) randomEXP() time.Time {
	if r.maxCacheDuration == r.minCacheDuration {
		return time.Now().Add(r.minCacheDuration)
	}
	exp := rand.Int63n(int64(r.maxCacheDuration-r.minCacheDuration) + r.minCacheDuration.Nanoseconds())
	return time.Now().Add(time.Duration(exp))
}

func newCacheEntry(ips []net.IP, exp time.Time) *cacheEntry {

	return &cacheEntry{
		ips:            ips,
		created:        time.Now(),
		expirationTime: exp,
	}
}

func newCacheEntryV2(ips []net.IP, exp time.Time) *cacheEntry {

	ipv4 := make([]net.IP, 0)
	ipv6 := make([]net.IP, 0)

	for _, ip := range ips {

		if ip.To4() != nil {
			ipv4 = append(ipv4, ip)
		} else {
			ipv6 = append(ipv6, ip)
		}

	}

	return &cacheEntry{
		ips: ips,
		//	ipsv4:          ipv4,
		//	ipsv6:          ipv6,
		created:        time.Now(),
		expirationTime: exp,
	}
}
