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

## Join us

Have questions, ideas, bug reports or just want to chat? Come join [our discord server](https://discord.com/channels/1240272304294985800/1311031796372344894).

## License and Code of Conduct

Code is under the [Apache License 2.0](LICENSE), documentation is [CC BY 4.0](LICENSE-documentation).

The SDC project is following the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/main/code-of-conduct.md). More information and links about the CNCF Code of Conduct are [here](https://www.cncf.io/conduct/).
