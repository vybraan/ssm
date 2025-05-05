// Copyright (c) 2025 Leonardo Faoro & authors
// SPDX-License-Identifier: BSD-3-Clause

// Package sshconf loads, parses SSH config files,
// tries to be thread-safe.
// ref: https://man.openbsd.org/ssh_config.5
package sshconf

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	som "github.com/thalesfsp/go-common-types/safeorderedmap"
)

type Config struct {
	// protects Hosts
	mu             sync.Mutex
	Hosts          []Host // higher priority
	secondaryHosts []Host // lower priority

	order Order
	path  string
}

type Host struct {
	Name    string
	Options *som.SafeOrderedMap[string]
}

// Order defines how hosts are organized when parsed.
type Order int

const (
	TagOrder Order = iota + 1
)

func New() *Config {
	return &Config{}
}

func (c *Config) SetOrder(o Order) {
	c.order = o
}

// Parse parses SSH config files from default known locations.
// User: ~/.ssh/config
// System: /etc/ssh/ssh_config
// Parse also follows `Include` statements via recursion.
func (c *Config) Parse() error {
	path, err := defaultConfigPath()
	if err != nil {
		return err
	}
	return c.parse(path)
}

// Parse parses SSH config file from custom location.
func (c *Config) ParsePath(s string) error {
	if !strings.HasPrefix(s, "/") {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		s = filepath.Join(wd, s)
	}
	return c.parse(s)
}

func (c *Config) GetHost(name string) Host {
	for _, h := range c.Hosts {
		if h.Name == name {
			return h
		}
	}
	return Host{}
}

func (c *Config) GetParamFor(host Host, key string) string {
	for _, h := range c.Hosts {
		if h.Name == host.Name {
			val, ok := h.Options.Get(key)
			if !ok {
				return ""
			}
			return val
		}
	}
	return ""
}

func (c *Config) GetPath() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.path
}

const (
	commentPrefix  = "#"
	tagPrefix      = "#tag:"
	tagOrderPrefix = "#tagorder"
)

func (c *Config) parse(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// clear hosts in case parse is
	// called multiple times.
	c.Hosts = []Host{}
	c.secondaryHosts = []Host{}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	c.path = path
	scanner := bufio.NewScanner(f)
	var tagOrder bool
	var currentHost *Host
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// set orderbyTag
		if line == tagOrderPrefix {
			tagOrder = true
		}
		if c.order == TagOrder {
			tagOrder = true
		}

		// ignore empty or comment line
		if line == "" ||
			strings.HasPrefix(line, commentPrefix) &&
				!strings.HasPrefix(line, tagPrefix) {
			continue
		}
		parts := strings.Fields(line)
		// malformed line, skip
		if len(parts) < 2 {
			continue
		}
		k, v := strings.ToLower(parts[0]), strings.Join(parts[1:], " ")
		// remove comment suffixes
		// when not a tag
		if !strings.HasPrefix(line, tagPrefix) {
			k = removeComments(k)
			v = removeComments(v)
		}
		// recurse include files
		if k == "include" {
			if !strings.HasPrefix(v, "/") {
				path = filepath.Dir(path)
				v = filepath.Join(path, v)
			}
			paths, err := filepath.Glob(v)
			if err != nil {
				return err
			}

			for _, path := range paths {
				cfg := New()
				err := cfg.parse(path) // recursion
				if err != nil {
					return err
				}
				c.Hosts = append(c.Hosts, cfg.Hosts...)
			}
		}
		// all blocks must start with Host key
		if k == "host" {
			if strings.Contains(v, "*") {
				continue
			}
			if currentHost != nil {
				newHost(tagOrder, currentHost, c)
			}
			currentHost = &Host{
				Name:    v,
				Options: som.New[string](),
			}
			continue
		}
		// if not a host key must be an option
		if currentHost != nil {
			currentHost.Options.Add(k, v)
		}
	}
	if currentHost != nil {
		newHost(tagOrder, currentHost, c)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	c.Hosts = append(c.Hosts, c.secondaryHosts...)
	return nil
}

func newHost(tagOrder bool, currentHost *Host, config *Config) {
	if tagOrder {
		if currentHost.Options.Contains("#tag:") {
			config.Hosts = append(config.Hosts, *currentHost)
		} else {
			config.secondaryHosts = append(config.secondaryHosts, *currentHost)
		}
		return
	}
	config.Hosts = append(config.Hosts, *currentHost)
}

func removeComments(input string) string {
	// find index of '#' and take substring up to that point
	if index := strings.Index(input, "#"); index != -1 {
		return strings.TrimSpace(input[:index])
	}
	// if no '#' found, return trimmed input
	return strings.TrimSpace(input)
}
