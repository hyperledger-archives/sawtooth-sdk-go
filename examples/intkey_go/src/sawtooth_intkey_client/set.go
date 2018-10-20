/**
 * Copyright 2018 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * ------------------------------------------------------------------------------
 */

package main

import (
	"github.com/jessevdk/go-flags"
	"strconv"
)

type Set struct {
	Args struct {
		Name  string `positional-arg-name:"name" required:"true" description:"Name of key to set"`
		Value string `positional-arg-name:"value" required:"true" description:"Amount to set"`
	} `positional-args:"true"`
	Url     string `long:"url" description:"Specify URL of REST API"`
	Keyfile string `long:"keyfile" description:"Identify file containing user's private key"`
	Wait    uint   `long:"wait" description:"Set time, in seconds, to wait for transaction to commit"`
}

func (args *Set) Name() string {
	return "set"
}

func (args *Set) KeyfilePassed() string {
	return args.Keyfile
}

func (args *Set) UrlPassed() string {
	return args.Url
}

func (args *Set) Register(parent *flags.Command) error {
	_, err := parent.AddCommand(args.Name(), "Sets an intkey value", "Sends an intkey transaction to set <name> to <value>.", args)
	if err != nil {
		return err
	}
	return nil
}

func (args *Set) Run() error {
	// Construct client
	name := args.Args.Name
	value, err := strconv.Atoi(args.Args.Value)
	if err != nil {
		return err
	}
	wait := args.Wait

	intkeyClient, err := GetClient(args, true)
	if err != nil {
		return err
	}
	_, err = intkeyClient.Set(name, uint(value), wait)
	return err
}
