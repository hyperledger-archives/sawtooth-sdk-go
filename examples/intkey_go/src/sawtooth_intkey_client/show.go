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
	"fmt"
	"github.com/jessevdk/go-flags"
)

type Show struct {
	Args struct {
		Name string `positional-arg-name:"name" required:"true" description:"Name of key to show"`
	} `positional-args:"true"`
	Url string `long:"url" description:"Specify URL of REST API"`
}

func (args *Show) Name() string {
	return "show"
}

func (args *Show) KeyfilePassed() string {
	return ""
}

func (args *Show) UrlPassed() string {
	return args.Url
}

func (args *Show) Register(parent *flags.Command) error {
	_, err := parent.AddCommand(args.Name(), "Displays the specified intkey value", "Shows the value of the key <name>.", args)
	if err != nil {
		return err
	}
	return nil
}

func (args *Show) Run() error {
	// Construct client
	name := args.Args.Name
	intkeyClient, err := GetClient(args, false)
	if err != nil {
		return err
	}
	value, err := intkeyClient.Show(name)
	if err != nil {
		return err
	}
	fmt.Println(name, ": ", value)
	return nil
}
