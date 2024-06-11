package main

import (
	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/anmitsu/go-shlex"
	"github.com/ergochat/readline"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"io"
	"os"

	"github.com/UsingCoding/gostore/data"
	"github.com/UsingCoding/gostore/internal/gostore/app/service"
	"github.com/UsingCoding/gostore/internal/gostore/app/store"
	"github.com/UsingCoding/gostore/internal/gostore/infrastructure/consoleoutput"
)

func repl(ctx *cli.Context) error {
	output := consoleoutput.New(
		os.Stdout,
		consoleoutput.WithNewline(true),
	)

	rl, err := readline.New("gostore> ")
	if err != nil {
		return err
	}
	defer func() {
		_ = rl.Close()
	}()

	// override ExitErrHandler to exclude application shutdown
	ctx.App.ExitErrHandler = func(c *cli.Context, err error) {
		if err == nil {
			return
		}
		output.Errorf("%s", err)
	}

	s, _ := newStoreService(ctx)

	data.Logo(output)
	output.Printf("Welcome to GoStore!")
	output.Printf("This is built-in shell with autocompletion")

	for {
		config := rl.GetConfig()
		autoCompleter, err2 := completer(ctx, s)
		if err2 != nil {
			return errors.Wrap(err2, "failed to build completer")
		}

		config.AutoComplete = autoCompleter

		err2 = rl.SetConfig(config)
		if err2 != nil {
			return errors.Wrap(err2, "failed to set readline config")
		}

		line, err2 := rl.ReadLine()
		if err2 != nil {
			if errors.Is(err2, readline.ErrInterrupt) || errors.Is(err2, io.EOF) {
				return nil
			}

			return errors.Wrap(err2, "failed to read line")
		}

		args, err2 := shlex.Split(line, false)
		if err2 != nil {
			return errors.Wrap(err2, "failed to split line")
		}

		if len(args) < 1 {
			continue
		}

		// ignore errors from app, since in REPL mode all errors from app displayed to user
		_ = ctx.App.RunContext(ctx.Context, append([]string{"gostore"}, args...))
	}
}

func completer(c *cli.Context, s service.Service) (readline.AutoCompleter, error) {
	recursivelyCollectArgs := func(cmd *cli.Command, args *[]readline.PrefixCompleterInterface) error {
		if cmd.Hidden {
			return nil
		}
		var subCommands []readline.PrefixCompleterInterface

		completions, err := completionItemForCmd(c, cmd.Name, s)
		if err != nil {
			return err
		}

		if len(completions) != 0 {
			subCommands = append(subCommands, completions...)
		}

		for _, subcommand := range cmd.Subcommands {
			subCommands = append(subCommands, readline.PcItem(subcommand.Name))
		}
		*args = append(*args, readline.PcItem(cmd.Name, subCommands...))
		for _, alias := range cmd.Aliases {
			*args = append(*args, readline.PcItem(alias, subCommands...))
		}

		return nil
	}

	var args []readline.PrefixCompleterInterface
	for _, command := range c.App.Commands {
		err := recursivelyCollectArgs(command, &args)
		if err != nil {
			return nil, err
		}
	}

	return readline.NewPrefixCompleter(args...), nil
}

func completionItemForCmd(c *cli.Context, cmd string, s service.Service) ([]readline.PrefixCompleterInterface, error) {
	switch cmd {
	case "get",
		"move",
		"copy",
		"remove":
		return storeEntriesPathsForCompletion(c, s)
	default:
		return nil, nil
	}
}

func storeEntriesPathsForCompletion(c *cli.Context, s service.Service) ([]readline.PrefixCompleterInterface, error) {
	tree, err := s.List(c.Context, store.ListParams{})
	if err != nil {
		return nil, err
	}

	return slices.Map(
		tree.Inline().Keys(),
		func(p string) readline.PrefixCompleterInterface {
			return readline.PcItem(p)
		},
	), nil
}
