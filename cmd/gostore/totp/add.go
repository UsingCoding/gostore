package totp

import (
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"

	clipkg "github.com/UsingCoding/gostore/internal/cli"
	"github.com/UsingCoding/gostore/internal/gostore/app/usecase/totp"
)

func add() *cli.Command {
	return &cli.Command{
		Name:  "add",
		Usage: "add a totp issuer",
		Action: func(c *cli.Context) error {
			prompt := promptui.Prompt{
				Label: "Name",
			}
			name, err := prompt.Run()
			if err != nil {
				return err
			}

			prompt = promptui.Prompt{
				Label: "Secret",
				Mask:  '*',
			}
			secret, err := prompt.Run()
			if err != nil {
				return err
			}

			prompt = promptui.Prompt{
				Label:   "ALG",
				Default: string(totp.AlgorithmSHA1),
			}
			alg, err := prompt.Run()
			if err != nil {
				return err
			}

			return clipkg.ContainerScope.MustGet(c.Context).TOTP.
				AddIssuer(
					c.Context,
					totp.AddParams{
						Name:      name,
						Secret:    []byte(secret),
						Algorithm: totp.Algorithm(alg),
					},
				)
		},
	}
}
