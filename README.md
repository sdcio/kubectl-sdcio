# kubectl-sdc

![sdc logo](https://docs.sdcio.dev/assets/logos/SDC-transparent-withname-100x133.png)

kubectl-sdc is the SDC specific kubectl plugin.

## notes
- Commands use the current kubectl config to access the cluster and namespace.
- `runningconfig` connects to `sdc-system/data-server` via port-forward.

## subcommands
kubectl-sdc provides the following functionalities.

### blame
The blame command provides a tree based view on the actual running device configuration of the given SDC target.

It takes the `--target` parameter, that defines which targets is to be displayed.
The `--format` parameter supports `tree` (default) and `xpath`; with `--format=xpath`, the `--interactive` flag opens a fuzzyfinder with multi-select (`Tab` to select) and prints the selected XPath lines.

For every configured attribute you will see the highes preference value as well as the source of that value.
- `running` are attributes that come from the device itself, where no intent exist in sdc.
- `default` is all the default values that are present in the config, that are not overwritten by any specific config.
- `<namespace>.<intentname>` is the reference to the intent that defined the actual highes preference value for that config attribute.
```
kubectl sdc blame --target srl1 --filter-owner running --format tree --filter-path /interface[name=mgmt0]/subinterface --filter-path /system/snmp/access-group[name=SNMPv2-RO-Community] --filter-owner default --filter-leaf admin*
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
kubectl sdc blame --target sros --filter-leaf "admin-state" --filter-owner "running"

# Show all interface-related configuration with deviations
kubectl sdc blame --target sros --filter-path "*/interface/*" --filter-deviation

# Show configuration from specific intent with timeout-related leaves
kubectl sdc blame --target sros --filter-owner "production.intent-emergency" --filter-leaf "*timeout*"

# Combine multiple filters to find specific configuration
kubectl sdc blame --target sros --filter-path "/config/service/emergency/*" --filter-leaf "ambulance" --filter-owner "test-system.*"
```

### runningconfig
The runningconfig command retrieves the running configuration for a target from the data-server.

It takes the `--target` parameter, that defines which target is to be displayed.
The `--format` parameter controls the output format (json, json_ietf, xml, xpath, yaml). Default is `xpath`.

Hints:
- The command uses the current kubectl config to access the cluster and namespace.
- The command connects to the `sdc-system/data-server` service (port-forward) to fetch the running config.

Example:
```
kubectl sdc runningconfig --target srl1 --format xpath
  
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
The deviation command lists and optionally reverts deviations.

By default, the command is **non-interactive**: it outputs all matching deviations directly.
Use `--interactive` to open the fuzzy finder selector.

It takes one of the following parameters:
- `--target`: shows all deviations for the specified target and limits `--deviation` autocompletion to deviations from that target.
- `--deviation`: shows only the specified deviation resource.

At least one of `--target` or `--deviation` must be provided.

Flags:
- `--format`: output format (`text` (default), `resource-yaml`, `resource-json`).
- `--filter-path`: filter deviation paths by prefix before selection/output. Can be repeated.
- `--revert`: clear the final selected/output deviations on the target.
- `--interactive`: enable interactive fuzzy finder mode.
- `--preview`: show preview panel in interactive mode.
- `--query`: initial fuzzy finder query in interactive mode.
- `--select-path-prefix`: mark matching path prefixes as selected in interactive mode. Can be repeated.
- `--auto-accept-select-path-prefix`: automatically confirm selected path prefixes in interactive mode.

`--revert` can be used with `--target`, `--deviation`, or both, and is compatible with `--preview`.

Mode behavior:
- Non-interactive (default): output all deviations after applying `--filter-path`.
- Interactive (`--interactive`): choose deviations in fuzzy finder; `--select-path-prefix` and `--query` apply here.
- `--select-path-prefix` and `--auto-accept-select-path-prefix` require `--interactive`.

Selection notes:
- The preview shows the actual and desired value for the currently selected path.
- When `--format=text`, exiting the interactive window prints the current human-readable deviation output.
- When `--format=resource-yaml` or `--format=resource-json`, exiting the interactive window prints a `TargetClearDeviation` manifest built from the selected entries.
- When `--revert` is set, the selected entries are cleared on the target.

Interactive quick keys:
- `Tab` toggle selection
- `Enter` confirm selection
- `Esc` / `Ctrl+C` abort
- `Ctrl+T` toggle preview
- `Ctrl+S` toggle selected-items view

Flag overview:

| Flag | Mode | Purpose |
|---|---|---|
| `--interactive` | Interactive | Open fuzzy finder selection UI |
| `--filter-path` | Both | Filter deviation paths by prefix before selection/output |
| `--select-path-prefix` | Interactive | Mark matching path prefixes as selected |
| `--auto-accept-select-path-prefix` | Interactive | Auto-confirm when `--select-path-prefix` matches |
| `--query` | Interactive | Seed fuzzy finder with an initial query |
| `--preview` | Interactive | Show details preview panel |
| `--revert` | Both | Clear final selected/output deviations on target |

Example (show all deviations for a target):
```bash
kubectl sdc deviation --target srl1
```

Example (non-interactive with path filtering):
```bash
kubectl sdc deviation --target srl1 --filter-path /interface --filter-path /system
```

Example (start the interactive view with an initial query):
```bash
kubectl sdc deviation --target srl1 --interactive --query interface
```

Example (interactive with selected path prefixes and auto-accept):
```bash
kubectl sdc deviation --target srl1 --interactive --select-path-prefix /interface --auto-accept-select-path-prefix
```

Example (render selected deviations as a `TargetClearDeviation` manifest):
```bash
kubectl sdc deviation --target srl1 --format resource-yaml
```

Example (interactive preview flow):
```bash
kubectl sdc deviation --deviation srl1 --interactive --preview
```

### apply
The apply command applies resources from YAML or JSON files, similar to kubectl apply.

Currently supported resource kinds:
- `TargetClearDeviation` (config.sdcio.dev/v1alpha1)

Input options:
- `-f`, `--filename`: one or more files to apply.
- Positional arguments are also accepted as file paths.
- Use `-` to read from stdin.

Behavior notes:
- Multi-document YAML files are supported.
- If a `TargetClearDeviation` manifest omits `metadata.namespace`, the current kubectl namespace is used.
- The command sends the resource to the target `cleardeviation` subresource.

Examples:

Apply from a file:
```bash
kubectl sdc apply -f clear-dev.yaml
```

Apply from multiple files:
```bash
kubectl sdc apply -f clear-dev-1.yaml -f clear-dev-2.yaml
```

Apply using positional file arguments:
```bash
kubectl sdc apply clear-dev.yaml another-dev.yaml
```

Apply from stdin:
```bash
cat clear-dev.yaml | kubectl sdc apply -f -
```

Example `TargetClearDeviation` manifest:
```yaml
apiVersion: config.sdcio.dev/v1alpha1
kind: TargetClearDeviation
metadata:
  name: srl1
spec:
  config:
    - name: intent-a
      paths:
        - /interface[name=ethernet-1/1]/admin-state
        - /system/name
```

## Join us

Have questions, ideas, bug reports or just want to chat? Come join [our discord server](https://discord.com/channels/1240272304294985800/1311031796372344894).

## License and Code of Conduct

Code is under the [Apache License 2.0](LICENSE), documentation is [CC BY 4.0](LICENSE-documentation).

The SDC project is following the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md). More information and links about the CNCF Code of Conduct are [here](https://www.cncf.io/conduct/).
