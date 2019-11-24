# Quick Start

Download Vigie and run tests
{: .subtitle }

!!! tip "Standalone Vigie"
    This quickstart will run Vigie in its simplest operating mode : _Standalone_.

## Get Vigie

### Download the binary

_Vigie is a single binary with no dependencies._

```bash tab="Linux"
wget -qO- https:// | tar  xvz
```

```bash tab="Windows"
wget -qO- https:// | tar  xvz ./vigie
```

### Get a Vigie configuration file

_Grab a pre-configure Vigie Config, ready to run._

!!! info "Vigie tests auto-provisioning"
    Tests will be downloaded by Vigie from the [vigie-demo-test](https://github.com/Vincoll/vigie-demo-test) git repo.

```bash tab="Linux"
wget -qO- https:// 
```

```bash tab="Windows"
wget -qO- https:// 
```

**Edit VigieConf (optional)**

You can quickly edit the config to configure the alerting, or change the API port if you wish.

## Run Vigie

```bash tab="Linux"
chmod +x ./vigie && \
vigie run --config vigieconf_getstarted.toml
```

```bash tab="Windows"
vigie run --config vigieconf_getstarted.toml
```

```bash tab="Linux Advanced"
chmod +x ./vigie && \
sudo setcap cap_net_raw=+ep && \
vigie run --config vigieconf_getstarted.toml
```

## Access the API
_No WebUI yet :/_

Go to [http://localhost:6680](http://localhost:6680)