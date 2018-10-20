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

type List struct {
	Url string `long:"url" description:"Specify URL of REST API"`
}

func (args *List) Name() string {
	return "list"
}

func (args *List) KeyfilePassed() string {
	return ""
}

func (args *List) UrlPassed() string {
	return args.Url
}

func (args *List) Register(parent *flags.Command) error {
	_, err := parent.AddCommand(args.Name(), "Displays all intkey values", "Shows the values of all keys in intkey state.", args)
	if err != nil {
		return err
	}
	return nil
}

func (args *List) Run() error {
	// Construct client
	intkeyClient, err := GetClient(args, false)
	if err != nil {
		return err
	}
	pairs, err := intkeyClient.List()
	if err != nil {
		return err
	}
	for _, pair := range pairs {
		for k, v := range pair {
			fmt.Println(k, ": ", v)
		}
	}
	return nil
}
