# kubectl-sdcio

![sdc logo](https://docs.sdcio.dev/assets/logos/SDC-transparent-withname-100x133.png)

kubectl-sdcio is the SDC specific kubectl plugin.

## notes
- Commands use the current kubectl config to access the cluster and namespace.
- Shell completion is available for `runningconfig` (`--target`, `--format`) and `deviation` (`--deviation`).
- `runningconfig` connects to `sdc-system/data-server` via port-forward.

## subcommands
kubectl-sdcio provides the following functionalities.

### blame
The blame command provides a tree based view on the actual running device configuration of the given SDC target.

It takes the `--target` parameter, that defines which targets is to be displayed.

For every configured attribute you will see the highes preference value as well as the source of that value.
- `running` are attributes that come from the device itself, where no intent exist in sdcio.
- `default` is all the default values that are present in the config, that are not overwritten by any specific config.
- `<namespace>.<intentname>` is the reference to the intent that defined the actual highes preference value for that config attribute.
```
kubectl sdcio blame --target srl1 --filter-owner running --format tree  --filter-path /interface[name=mgmt0]/subinterface --filter-path /system/snmp/access-group[name=SNMPv2-RO-Community] --filter-owner default --filter-leaf admin*
  -----    │     🎯 default.srl1
  -----    │     ├── 📦 interface
  -----    │     │   └── 🔑 name=mgmt0
  -----    │     │       └── 📦 subinterface
  -----    │     │           └── 🔑 index=0
running    │     │               ├── 🍃 admin-state -> enable
  -----    │     │               ├── 📦 ipv4
running    │     │               │   └── 🍃 admin-state -> enable
  -----    │     │               └── 📦 ipv6
running    │     │                   └── 🍃 admin-state -> enable
  -----    │     └── 📦 system
  -----    │         └── 📦 snmp
  -----    │             └── 📦 access-group
  -----    │                 └── 🔑 name=SNMPv2-RO-Community
default    │                     └── 🍃 admin-state -> enable
```

#### Filtering Options

The blame command supports several filtering options to narrow down the results. **All filters are cumulative** (combined with "AND" logic), meaning only configuration elements that match ALL specified criteria will be displayed.

Available filters:

- **`--filter-leaf <pattern>`**: Filter by leaf node name. Supports wildcards (`*`).
  - Example: `--filter-leaf "admin-state"` shows only admin-state leaves
  - Example: `--filter-leaf "interface*"` shows all leaves starting with "interface"

- **`--filter-owner <pattern>`**: Filter by configuration owner. Supports wildcards (`*`).
  - Example: `--filter-owner "running"` shows only running configuration
  - Example: `--filter-owner "default.*"` shows all default configurations
  - Example: `--filter-owner "production.intent-*"` shows intents from production namespace

- **`--filter-path <pattern>`**: Filter by configuration path. Supports wildcards (`*`).
  - Example: `--filter-path "/config/service/*"` shows only service-related configuration
  - Example: `--filter-path "*/interface/*"` shows interface configuration at any level
The whole path (including leaves) is involved in the pattern matching.

- **`--filter-deviation`**: Show only configuration elements that have deviations between intended and actual values.

#### Filter Examples

```bash
# Show only admin-state configuration from running config
kubectl sdcio blame --target sros --filter-leaf "admin-state" --filter-owner "running"

# Show all interface-related configuration with deviations
kubectl sdcio blame --target sros --filter-path "*/interface/*" --deviation

# Show configuration from specific intent with timeout-related leaves
kubectl sdcio blame --target sros --filter-owner "production.intent-emergency" --filter-leaf "*timeout*"

# Combine multiple filters to find specific configuration
kubectl sdcio blame --target sros --filter-path "/config/service/emergency/*" --filter-leaf "ambulance" --filter-owner "test-system.*"


### runningconfig
The runningconfig command retrieves the running configuration for a target from the data-server.

It takes the `--target` parameter, that defines which target is to be displayed.
The `--format` parameter controls the output format (json, json_ietf, xml, xpath, yaml). Default is `xpath`.

Hints:
- The command uses the current kubectl config to access the cluster and namespace.
- The command connects to the `sdc-system/data-server` service (port-forward) to fetch the running config.
- Shell completion is available for `--target` and `--format`.

Example:
```
kubectl sdcio runningconfig --target srl1 --format xpath
  
/system/snmp/access-group[name=SNMPv2-RO-Community]/name: SNMPv2-RO-Community
/system/snmp/access-group[name=SNMPv2-RO-Community]/security-level: no-auth-no-priv
/system/snmp/network-instance[name=mgmt]/admin-state: enable
/system/snmp/network-instance[name=mgmt]/name: mgmt
/system/ssh-server[name=mgmt-netconf]/admin-state: enable
/system/ssh-server[name=mgmt-netconf]/disable-shell: true
/system/ssh-server[name=mgmt-netconf]/name: mgmt-netconf
/system/ssh-server[name=mgmt-netconf]/network-instance: mgmt
/system/ssh-server[name=mgmt-netconf]/port: 830
/system/ssh-server[name=mgmt]/admin-state: enable
/system/ssh-server[name=mgmt]/name: mgmt
/system/ssh-server[name=mgmt]/network-instance: mgmt
/system/ssh-server[name=mgmt]/use-credentialz: true
/system/tls/server-profile[name=clab-profile]/authenticate-client: false
...
```

### deviation
The deviation command provides an xpath based view of the defined deviations. Deviations are auto completed from the k8s resources and can be previewed with details.

It takes the `--deviation` parameter, that defines which deviation is to be displayed.
The `--preview` flag toggles the details preview panel for the currently selected path.

Selection notes:
- The preview shows the actual and desired value for the currently selected path.
- When you exit the interactive window, the selected paths are printed to stdout.

Keyboard shortcuts:
- Tab selects entries (multi selection is supported).
- Shift + Left/Right Arrow horizontally scrolls the list.
- Ctrl + O toggles searching for paths only vs paths and values.

Example (preview a deviation):
```
kubectl sdcio deviation --deviation srl1 --preview

  Namespace: default, Deviation: srl1 [target]                                                           ┌───────────────────────────────────────────────────────────────────────────────────────────────────────┐
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=110]/description                            │ Path:    /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=1000]/description                     │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=110]/action/accept                          │ Actual:  Drop all else                                                                                │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=10]/sequence-id                             │ Desired:                                                                                              │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=10]/match/ipv4/protocol                     │ Reason:  UNHANDLED                                                                                    │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=10]/match/ipv4/icmp/type                    │                                                                                                       │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=10]/description                             │                                                                                                       │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=10]/action/accept/rate-limit/system-cpu-p.. │                                                                                                       │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=100]/sequence-id                            │                                                                                                       │
 >[U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=100]/match/transport/destination-port/val.. │                                                                                                       │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=100]/match/transport/destination-port/ope.. │                                                                                                       │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=100]/match/ipv4/protocol                    │                                                                                                       │
 >[U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=100]/description                            │                                                                                                       │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=100]/action/accept                          │                                                                                                       │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=1000]/sequence-id                           │                                                                                                       │
>>[U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=1000]/description                           │                                                                                                       │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=1000]/action/log                            │                                                                                                       │
  [U] /acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=1000]/action/drop                           │                                                                                                       │
  739/739                                                                                                │                                                                                                       │
>                                                                                                        └───────────────────────────────────────────────────────────────────────────────────────────────────────┘

# Output after exit (stdout)
/acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=100]/match/transport/destination-port/value
/acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=100]/description
/acl/acl-filter[name=cpm][type=ipv4]/entry[sequence-id=1000]/description
```

## Join us

Have questions, ideas, bug reports or just want to chat? Come join [our discord server](https://discord.com/channels/1240272304294985800/1311031796372344894).

## License and Code of Conduct

Code is under the [Apache License 2.0](LICENSE), documentation is [CC BY 4.0](LICENSE-documentation).

The SDC project is following the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md). More information and links about the CNCF Code of Conduct are [here](https://www.cncf.io/conduct/).
