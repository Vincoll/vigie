# Vigie Modes

Vigie can be launched in several operating modes. Switching from one mode to another is done by activating features in the config file.

!!! tip "Reminder"
    Bear in mind that a monitoring solution must be more reliable than what you want to monitor.


### Standalone
*Vigie*
Single instance of Vigie, no other software dependencies.
Alerting is present, but the functionalities are minimal.

#### Great for:
  * Simple monitoring of non-critical services
  * Low footprint

### Complete Simple
*Vigie + TSDB ([InfluxDB](https://www.influxdata.com/) or [Warp10](https://www.warp10.io/) *(WIP)*)*


An instance of Vigie is coupled to a TimeSeries database. Each test result is written in a TimeSeries DB.
This allows to save every the test returns. Grafana dashboards will allow you to analyze the results. More advanced alerting can be managed throw your TSDB.

#### Great for:
  * Monitoring non-critical services
  * Audit
  * Observability
  * Report SLO
 
### Complete High Availability
*Vigies + [Consul](https://www.consul.io/) + HA TSDB ([InfluxDB](https://www.influxdata.com/) or [Warp10](https://www.warp10.io/) *(WIP)*)*

Multiples instances of Vigie can be deployed, one of them will be a leader, others followers. The testsuite will be distributed on all the running Vigie instances and rescheduled as soon as the number of instances changes.\
The overhall Vigie state is saved into a [Consul](https://www.consul.io/) cluster.\
Each test result is written in a TimeSeries DB.
This allows to save every the test returns. Grafana dashboards will allow you to analyze the results. More advanced alerting can be managed throw your TSDB.

#### Great for:
  * Monitoring critical services
  * Scaling
  * Audit
  * Observability
  * Report SLO

### Fully Distributed
*Future* 