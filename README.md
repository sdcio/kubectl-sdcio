# kubectl-sdcio

![sdc logo](https://docs.sdcio.dev/assets/logos/SDC-transparent-withname-100x133.png)

kubectl-sdcio is the SDC specific kubectl plugin.

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

## Join us

Have questions, ideas, bug reports or just want to chat? Come join [our discord server](https://discord.com/channels/1240272304294985800/1311031796372344894).

## License and Code of Conduct

Code is under the [Apache License 2.0](LICENSE), documentation is [CC BY 4.0](LICENSE-documentation).

The SDC project is following the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md). More information and links about the CNCF Code of Conduct are [here](https://www.cncf.io/conduct/).
