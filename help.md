NAME:
   ssm - Secure Shell Manager

USAGE:
   ssm [--options] [tag]
   example: ssm --show --exit vpn
   example: ssm -se vpn

VERSION:
   0.2.1
   build date: 2025-04-26T22:15:58Z
   build SHA: eabcd3c5458e611a8600d448a438cdc9be1f03de

DESCRIPTION:
   SSM is an open source (MIT) SSH connection manager that helps engineers organize servers, connect, filter, tag, execute commands (soon), transfer files (soon), and much more from a simple terminal interface.

AUTHOR:
   "Leonardo Faoro" <ssm@leonardofaoro.com>

GLOBAL OPTIONS:
   --show, -s                  always show config (default: false)
   --exit, -e                  exit after connection (default: false)
   --config string, -c string  custom config file path
   --debug, -d                 enable debug mode with verbose logging (default: false)
   --help, -h                  show help
   --version, -v               print the version

COPYRIGHT:
   (c) Leonardo Faoro (MIT)
