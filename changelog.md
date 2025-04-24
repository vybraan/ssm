# [0.2.0-next]
- add exit flag `--exit / -e`: ssm will exit after connecting to a host
- add `ctrl+v`: view full config for selected host

# [0.1.2] - April 21, 2025
- fix parsing of tag keys

# [0.1.1] - April 21, 2025
- fix parsing comments on same line as config keys
- move segfault free server at the bottom
- resolve absolute path from custom --config
- add help section to readme
- add ssh config example in data/config_example

# [0.1.0] - April 20, 2025
- extend pkg/sshconf to support #tag: keys e.g. #tag: admin,vpn
- add arg for tags e.g. `ssm admin` will show only admin tagged hosts
- add `--config, -c` flag to provide custom config location other than default search paths

# [0.0.1] - April 18, 2025
- initial release
- pkg/sshconf: parse, watch logic 
- pkg/tui: bubbletea UI implementation
- main.go: initilization logic, args & flags handling

[0.0.1]: https://github.com/lfaoro/ssm/releases/tag/0.0.1
[0.1.0]: https://github.com/lfaoro/ssm/compare/0.0.1...0.1.0
[0.1.1]: https://github.com/lfaoro/ssm/compare/0.1.0...0.1.1
[0.1.2]: https://github.com/lfaoro/ssm/compare/0.1.1...0.1.2
[0.2.0]: https://github.com/lfaoro/ssm/compare/0.1.2...0.2.0
