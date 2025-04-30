NAME:
   ssm - Secure Shell Manager

USAGE:
   ssm [--options] [tag]
   example: ssm --show --exit vpn
   example: ssm -se vpn

VERSION:
   0.3.0

DESCRIPTION:
   SSM is an open source (MIT) SSH connection manager that helps engineers organize servers, connect, filter, tag, execute commands (soon), transfer files (soon), and much more from a simple terminal interface.

AUTHOR:
   "Leonardo Faoro" <ssm@leonardofaoro.com>

GLOBAL OPTIONS:
   --show, -s                  always show config params (default: false)
   --exit, -e                  exit after connection (default: false)
   --config string, -c string  custom config file path
   --debug, -d                 enable debug mode with verbose logging (default: false)
   --help, -h                  show help
   --version, -v               print the version

COPYRIGHT:
   (c) Leonardo Faoro (MIT)
