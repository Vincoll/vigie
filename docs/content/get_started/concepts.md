# Concepts

Vigie's concepts in a nutshell. Find the details in their respective sections.
{: .subtitle }

## Tests

The organization of the Tests is in 3 parts. 

File > TestSuite > []TestCase > []Teststep

**Structure**

Example of a TestSuit with one TestCase, One TestStep

```yaml
name: TestSuite Vigie

config:
  frequency:
    http: 3m
  timeout:
    http: 5s

vars:
  google: ["google.fr,google.com"]

testcases:

- name: Example HTTP

  steps:

    - name: GET examples
      probe:
        type: http
        follow_redirects: true
        method: GET
        url: $item
      assertions:
        - probeinfo.responsetime < 600ms
        - httpcode == 200
      loop:
        - $google
        - foo.tld

```

**config**

Each Test* have a `config` property design to tweak your monitoring frequency. 

```yaml
config:
  frequency:
    dns: 120s
  timeout:
    dns: 1s

```
Each one inherits from the other, it's the results of the leaves that count.

**loop**  

A loop allows you to multiply the TestStep by the number of elements included in a list. 


## Probes

Vigie has several types of probes covering different protocols. Each one takes different parameters, thus making it possible to test 
all kinds of combinations.


Exemples:

```yaml
- probe:
    type: dns
    FQDN: "txt3.dns.test.vigie.dev."
    RecordType: "TXT"
```
```yaml
- probe:
    type: icmp
    ipversion: 4
    host: vigie.dev
```
```yaml
- probe:
    type: x509
    host: vigie.dev
    port: 443
```

For each probe the application response is testable according to the expected result.
Follow the documentation of each Probe in order to know its capabilities.  

Note: In case a Domain Name returns multiple IP addresses, each of these addresses will be tested and filled in a *`subtest`*.


## Assertions

Following the return of a probe, the assertions allow the response to be tested in depth.

Multiple assertion verbs cover the majority of cases. (equality, difference, greater than, contained in)

> Documentation complete Assertions

Exemples:

```yaml
- probe:
  type: x509
  host: abc.xyz
  port: 443
assertions:
  - valid == true
  - daybeforeexpiration > 15
  - expired == false
  - endcertificate.dnsnames $$ "*.golang.com"
  - rootcertificate.publickeyalgorithm != "3DES"
  - endcertificate.signaturealgorithm == "SHA256-RSA"
```