# kubectl-sdcio

kubectl-sdcio is the sdcio specific kubectl plugin.

## subcommands
kubectl-sdcio provides the following functionalities.

### blame
The blame command provides a tree based view on the actual running device configuration of the given sdcio target.

It takes the `--target` parameter, that defines which targets is to be displayed.

For every configured attribute you will see the highes preference value as well as the source of that value.
- `running` are attributes that come from the device itself, where no intent exist in sdcio.
- `default` is all the default values that are present in the config, that are not overwritten by any specific config.
- `<namespace>.<intentname>` is the reference to the intent that defined the actual highes preference value for that config attribute.
```
mava@server01:~/projects/kubectl-sdcio$ kubectl sdcio blame --target sros
                    -----    │     🎯 default.sros
                    -----    │     └── 📦 configure
                    -----    │         ├── 📦 card
                    -----    │         │   └── 📦 1
                  default    │         │       ├── 🍃 admin-state -> enable
                  running    │         │       ├── 🍃 card-type -> iom-1
                  default    │         │       ├── 🍃 fail-on-error -> false
                  default    │         │       ├── 🍃 filter-profile -> none
                  default    │         │       ├── 🍃 hash-seed-shift -> 2
                  default    │         │       ├── 🍃 power-save -> false
                  default    │         │       ├── 🍃 reset-on-recoverable-error -> false
                  running    │         │       └── 🍃 slot-number -> 1
                    -----    │         ├── 📦 service
                    -----    │         │   ├── 📦 customer
                    -----    │         │   │   ├── 📦 1
    default.customer-sros    │         │   │   │   ├── 🍃 customer-id -> 1
    default.customer-sros    │         │   │   │   └── 🍃 customer-name -> 1
                    -----    │         │   │   └── 📦 2
    default.customer-sros    │         │   │       ├── 🍃 customer-id -> 2
    default.customer-sros    │         │   │       └── 🍃 customer-name -> 2
                    -----    │         │   └── 📦 vprn
                    -----    │         │       ├── 📦 vprn123
default.intent1-sros-sros    │         │       │   ├── 🍃 admin-state -> enable
                  default    │         │       │   ├── 🍃 allow-export-bgp-vpn -> false
                  default    │         │       │   ├── 🍃 carrier-carrier-vpn -> false
                  default    │         │       │   ├── 🍃 class-forwarding -> false
default.intent1-sros-sros    │         │       │   ├── 🍃 customer -> 1
...
```