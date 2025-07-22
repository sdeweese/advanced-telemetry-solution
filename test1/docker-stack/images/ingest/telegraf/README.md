# Telegraf join operation

This image adds custom plugins to the official telegraf image that allow us to
perform join operations on incoming metrics in the `sys/intf` measurement. Join
operations are necessary to tag all metrics for a given interface with important
metadata like interface description, administrative status and operational
status.

Here is an example of how to use the native and custom plugins to perform a join
operation on incoming interface metrics.

### Step 1: Ingest metrics from the device

The Nexus devices are configured to send interface telemetry metrics for a
select list of DME classes.

```bash
sensor-group interfaces
    path sys/intf query-condition query-target=subtree&target-subtree-class=l1PhysIf,ethpmPhysIf,pcAggrIf,ethpmAggrIf,rmonIfIn,rmonIfOut,ethpmFcot,eqptFcotSensor
```

#### Metadata

- `l1PhysIf`
  - The object that represents the Layer 1 physical Ethernet interface
    information object.
- `ethpmPhysIf`
  - Physical interface information holder
- `pcAggrIf`
  - The aggregated interface, which is a collection of physical ports; aka port
    channel
- `ethpmAggrIf`
  - Port channel interface information.
- `ethpmFcot`
  - Class for regular Fibre Channel optical transmitter types.

#### Measurements

- `rmonIfIn`
  - The interface input statistics.
- `rmonIfOut`
  - The interface output statistics.
- `eqptFcotSensor`
  - The transceiver DOM sensor information

The data arrives at telegraf in the following format. I am omitting many fields
and classes for brevity.

```bash
sys/intf switch=602-es09-tor1,dn=sys/intf/phys-[eth1/31]/phys descr="distributel_a_1",adminSt=up,operSt=up,... 1626357742000000000
sys/intf switch=602-es09-tor1,dn=sys/intf/phys-[eth1/31]/dbgIfIn octetRate=0,errors=0,discards=0,... 1626357742000000000
sys/intf switch=602-es09-tor1,dn=sys/intf/phys-[eth1/31]/dbgIfOut octetRate=0,errors=0,discards=0,... 1626357742000000000
sys/intf switch=602-es09-tor1,dn=sys/intf/phys-[eth1/32]/phys descr="distributel_a_2",adminSt=up,operSt=up,... 1626357742000000000
sys/intf switch=602-es09-tor1,dn=sys/intf/phys-[eth1/32]/dbgIfIn octetRate=0,errors=0,discards=0,... 1626357742000000000
sys/intf switch=602-es09-tor1,dn=sys/intf/phys-[eth1/32]/dbgIfOut octetRate=0,errors=0,discards=0,... 1626357742000000000
```

The above metrics are examples of how telegraf receives `sys/intf` metrics from
a given device. Each row represents the metrics for a DME path on the device.
The first 3 rows are metrics for the interface `eth1/31` and the last 3 rows are
metrics for the interface `eth1/32`. Notice that the first row contains metadata
while the second and third rows contain measurement data. Our goal is to tag the
measurement data in the second and third rows with the metadata in the first
row. To accomplish this, we will need to first be able to group measurements
together by their respective interface IDs.

### Step 2: Group measurements by interface ID

```toml
[[processors.regex]]
  order = 1
  namepass = ["sys/intf"]

  [[processors.regex.tags]]
    ## Regular expression to match on a tag name
    key = "dn"
    ## example dn: "sys/intf/phys-[eth1/42]"
    pattern = '^sys\/intf\/(phys|aggr)-\[(.*)\].*$'
    ## Replacement expression defining the name of the new tag
    replacement = "${2}"
    ## example result_key: "eth1/42"
    result_key = "interface"

  [processors.regex.tagdrop]
    run_time = ["after_merge"]

[[processors.converter]]
  order = 2
  namepass = ["sys/intf"]
  [processors.converter.tags]
    string = ["dn"]
  [processors.converter.tagdrop]
    run_time = ["after_merge"]
```

```
sys/intf switch=602-es09-tor1,interface=eth1/31 dn=sys/intf/phys-[eth1/31]/phys,descr="distributel_a_1",adminSt=up,operSt=up,... 1626357742000000000
sys/intf switch=602-es09-tor1,interface=eth1/31 dn=sys/intf/phys-[eth1/31]/dbgIfIn,octetRate=0,errors=0,discards=0,... 1626357742000000000
sys/intf switch=602-es09-tor1,interface=eth1/31 dn=sys/intf/phys-[eth1/31]/dbgIfOut,octetRate=0,errors=0,discards=0,... 1626357742000000000
sys/intf switch=602-es09-tor1,interface=eth1/32 dn=sys/intf/phys-[eth1/32]/phys,descr="distributel_a_2",adminSt=up,operSt=up,... 1626357742000000000
sys/intf switch=602-es09-tor1,interface=eth1/32 dn=sys/intf/phys-[eth1/32]/dbgIfIn,octetRate=0,errors=0,discards=0,... 1626357742000000000
sys/intf switch=602-es09-tor1,interface=eth1/32 dn=sys/intf/phys-[eth1/32]/dbgIfOut,octetRate=0,errors=0,discards=0,... 1626357742000000000
```

The above metrics have a new tag `interface` that is derived from `dn`. This
tag, along with the `switch` tag will be used to perform the join operation. The
`dn` tag has also been demoted to a field. Note the first three rows all have
the same key and can therefore be grouped together. However, there is a problem:
the fields have overlapping names. `octetRate` appears in both the second and
third rows. If we group the first three rows together, we will lose one of those
fields and will be unable to distinguish if it is an incoming or outgoing
metric. This is where our custom `tag_prefix` plugin comes in.

### Step 3: Prefix field names with their dn value and merge

```toml
[[processors.execd]]
  alias = "tag_prefix"
  order = 3
  namepass = ["sys/intf"]
  command = ["/usr/bin/tag_prefix"]
  [processors.execd.tagdrop]
    run_time = ["after_merge"]

# Merge metrics into multifield metrics by series key
[[aggregators.merge]]
  namepass = ["sys/intf"]
  ## Precision to round the metric timestamp to
  ## This is useful for cases where metrics to merge arrive within a small
  ## interval and thus vary in timestamp. The timestamp of the resulting metric
  ## is also rounded.
  # round_timestamp_to = "1ns"
  drop_original = true
  [aggregators.merge.tags]
    run_time = "after_merge"

```

```bash
sys/intf switch=602-es09-tor1,interface=eth1/31 sys/intf/phys-[eth1/31]/phys|dn=sys/intf/phys-[eth1/31]/phys,sys/intf/phys-[eth1/31]/phys|descr="distributel_a_1",sys/intf/phys-[eth1/31]/phys|adminSt=up,operSt=up,sys/intf/phys-[eth1/31]/dbgIfIn|octetRate=0,sys/intf/phys-[eth1/31]/dbgIfIn|errors=0,sys/intf/phys-[eth1/31]/dbgIfIn|discards=0,sys/intf/phys-[eth1/31]/dbgIfIn|octetRate=0,sys/intf/phys-[eth1/31]/dbgIfIn|errors=0,sys/intf/phys-[eth1/31]/dbgIfIn|discards=0,... 1626357742000000000
sys/intf switch=602-es09-tor1,interface=eth1/32 sys/intf/phys-[eth1/32]/phys|dn=sys/intf/phys-[eth1/32]/phys,sys/intf/phys-[eth1/32]/phys|descr="distributel_a_2",sys/intf/phys-[eth1/32]/phys|adminSt=up,operSt=up,sys/intf/phys-[eth1/32]/dbgIfIn|octetRate=0,sys/intf/phys-[eth1/32]/dbgIfIn|errors=0,sys/intf/phys-[eth1/32]/dbgIfIn|discards=0,sys/intf/phys-[eth1/32]/dbgIfIn|octetRate=0,sys/intf/phys-[eth1/32]/dbgIfIn|errors=0,sys/intf/phys-[eth1/32]/dbgIfIn|discards=0,... 1626357742000000000
```

Now the 6 rows have been reduced to 2 rows, one for each interface. The fields
have been prefixed with their `dn` value to avoid overlapping field names. The
`tag_prefix` plugin is a custom plugin that prefixes fields with the value of
the `dn` tag. The `merge` aggregator is used to merge the metrics with identical
keys and timestamps into a single row. We can now promote the metadata fields to
tags and then split the measurements back into separate rows using another
custom plugin called `split_metrics`.

### Step 4: Promote metadata fields to tags and split metrics

```toml
[[processors.converter]]
  order = 5
  namepass = ["sys/intf"]
  [processors.converter.fields]
    tag = [
    "sys/intf/phys-*|descr",
    "sys/intf/aggr-*|descr",
    "sys/intf/phys-*/phys|adminSt",
    "sys/intf/aggr-*/aggrif|adminSt",
    "sys/intf/phys-*/phys|operSt",
    "sys/intf/aggr-*/aggrif|operSt",
    ]

  [processors.converter.tagpass]
    run_time = ["after_merge"]

[[processors.regex]]
  order = 6
  namepass = ["sys/intf"]
  [[processors.regex.tag_rename]]
    pattern = '^sys\/intf\/(phys|aggr)-\[.*\]\|descr$'
    replacement = "descr"
  [[processors.regex.tag_rename]]
    pattern = '^sys\/intf\/(phys|aggr)-\[.*\]\/(phys|aggrif)\|adminSt$'
    replacement = "adminSt"
  [[processors.regex.tag_rename]]
    pattern = '^sys\/intf\/(phys|aggr)-\[.*\]\/(phys|aggrif)\|operSt$'
    replacement = "operSt"
  [processors.regex.tagpass]
    run_time = ["after_merge"]

[[processors.execd]]
  alias = "split_metrics"
  order = 7
  namepass = ["sys/intf"]
  command = ["/usr/bin/split_metrics"]
  [processors.execd.tagpass]
    run_time = ["after_merge"]
```

```bash
sys/intf switch=602-es09-tor1,interface=eth1/31,descr="distributel_a_1",adminSt=up,operSt=up,dn=sys/intf/phys-[eth1/31]/dbgIfIn octetRate=0,errors=0,discards=0,... 1626357742000000000
sys/intf switch=602-es09-tor1,interface=eth1/31,descr="distributel_a_1",adminSt=up,operSt=up,dn=sys/intf/phys-[eth1/31]/dbgIfOut octetRate=0,errors=0,discards=0,... 1626357742000000000
sys/intf switch=602-es09-tor1,interface=eth1/32,descr="distributel_a_2",adminSt=up,operSt=up,dn=sys/intf/phys-[eth1/32]/dbgIfIn octetRate=0,errors=0,discards=0,... 1626357742000000000
sys/intf switch=602-es09-tor1,interface=eth1/32,descr="distributel_a_2",adminSt=up,operSt=up,dn=sys/intf/phys-[eth1/32]/dbgIfOut octetRate=0,errors=0,discards=0,... 1626357742000000000
```

The `descr`, `adminSt` and `operSt` fields have been promoted to tags and
renamed. The `split_metrics` plugin is a custom plugin that splits the metrics
back into separate rows based on the the prefix of the field names. It also adds
bach the `dn` tag. Now we have achieved our goal of tagging the measurement data
with the metadata.
