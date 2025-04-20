# [0.1.0] - April 21, 2025
- extend pkg/sshconf to support #tag: keys e.g. #tag: admin,vpn
- add arg for tags e.g. `ssm admin` will show only admin tagged hosts
- add `--config, -c` flag to provide custom config location other than default search paths

# [0.0.1] - April 18, 2025
- initial release
- pkg/sshconf: parse, watch logic 
- pkg/tui: bubbletea UI implementation
- main.go: initilization logic, args & flags handling

[0.1.0]: https://github.com/lfaoro/ssm/compare/0.0.1...0.1.0
[0.0.1]: https://github.com/lfaoro/ssm/releases/tag/0.0.1
