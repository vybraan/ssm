# [0.3.5] next
- add show comments for keys in view mode
- add run command view

# [0.3.4] May 9, 2025
- resize list dynamically when error
- add ctrl+c to quit the app
- remove segfault auto add

# [0.3.3] May 5, 2025
- add #tagorder key to show `#tag` hosts first
- add `--order` flag to show `#tag` hosts first
- add clear filter using `backspace`
- fix could not resolve hostname bug
- fix show Host when HostName is missing
- refactor ssh config parser
- add when EDITOR not set search for vim,vi,nano,ed

# [0.3.2] April 30, 2025
- fix error/log message alignment
- add emacs keys: ctrl+p/n/b/f(up/down/left/right)
- update libraries

# [0.3.1] April 30, 2025
- fix exithost invalid character ssh error
- exit filtering on enter key if prompt is empty
- remove watcher from parser
- skip parsing wildcard(*) hosts

# [0.3.0] April 30, 2025
- add cursor while filtering
- restructure codebase

# [0.2.2] April 29, 2025
- return error when EDITOR env is not set
- add version check
- inform via msg when new version is released
- fix custom config path
- improve codebase

# [0.2.1] April 25, 2025
- fix crash on segfault
- remove windows release
- upgrade deps
- general improvements

# [0.2.0] April 24, 2025
- add exit flag `--exit / -e`: ssm will exit after connecting to a host
- add `ctrl+v`: view full config for selected host
- add ordered map for config options
- fix filtering hosts
- improve cli helpfile
- improve readme

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
[0.2.1]: https://github.com/lfaoro/ssm/compare/0.2.0...0.2.1
[0.2.2]: https://github.com/lfaoro/ssm/compare/0.2.1...0.2.2
[0.3.0]: https://github.com/lfaoro/ssm/compare/0.2.2...0.3.0
[0.3.1]: https://github.com/lfaoro/ssm/compare/0.3.0...0.3.1
[0.3.2]: https://github.com/lfaoro/ssm/compare/0.3.1...0.3.2
[0.3.3]: https://github.com/lfaoro/ssm/compare/0.3.2...0.3.3
[0.3.4]: https://github.com/lfaoro/ssm/compare/0.3.3...0.3.4
