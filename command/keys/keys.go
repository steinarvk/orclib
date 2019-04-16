package keys

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/steinarvk/orc"
	"github.com/steinarvk/orclib/lib/cryptopacket"
	"github.com/steinarvk/orclib/lib/mutatefile"
	"github.com/steinarvk/orclib/lib/orckeys"

	persistentkeys "github.com/steinarvk/orclib/module/orc-persistentkeys"
	publickeyregistry "github.com/steinarvk/orclib/module/orc-publickeyregistry"
	orctinkgcpkms "github.com/steinarvk/orclib/module/orc-tinkgcpkms"
)

var (
	KeysCommand = orc.Command(nil, nil, cobra.Command{
		Use:   "keys",
		Short: "Commands to manipulate cryptographic keys",
	}, nil)
)

func init() {
	var masterKeyURI string
	masterKeyURIFlags := orc.FlagsModule(func(flags *pflag.FlagSet) {
		flags.StringVar(&masterKeyURI, "master_key_uri", "", "URI to master key for keys")
	})

	var recipientFlag string
	recipientFlags := orc.FlagsModule(func(flags *pflag.FlagSet) {
		flags.StringVar(&recipientFlag, "recipient", "", "intended recipient of message")
	})

	var keyOwnerFlag string
	keyOwnerFlags := orc.FlagsModule(func(flags *pflag.FlagSet) {
		flags.StringVar(&keyOwnerFlag, "canonical_host", "", "canonical host (or name) of owner of the keys")
	})

	var registryToUpdateFilename string
	registryToUpdateFlags := orc.FlagsModule(func(flags *pflag.FlagSet) {
		flags.StringVar(&registryToUpdateFilename, "update_public_keys", "", "public key registry file to update")
	})

	maybeUpdateRegistry := func(public orckeys.PublicKeyPacket) error {
		if registryToUpdateFilename == "" {
			return nil
		}
		var existing map[string]orckeys.PublicKeyPacket
		return mutatefile.MutateFile(registryToUpdateFilename, 0600, func(data []byte) ([]byte, error) {
			if len(data) == 0 {
				existing = map[string]orckeys.PublicKeyPacket{}
			} else {
				if err := json.Unmarshal(data, &existing); err != nil {
					return nil, fmt.Errorf("Invalid data: %v", err)
				}
			}
			existing[public.Metadata.Owner] = public
			return json.Marshal(existing)
		})
	}

	orc.Command(KeysCommand, orc.Modules(
		persistentkeys.M,
		registryToUpdateFlags,
	), cobra.Command{
		Use:   "print-public",
		Short: "Print the public part of a set of keys",
	}, func() error {
		if err := maybeUpdateRegistry(persistentkeys.M.Keys.Public()); err != nil {
			return err
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(persistentkeys.M.Keys.Public())

		return nil
	})

	generateFlags := orc.Modules(
		masterKeyURIFlags,
		keyOwnerFlags,
		registryToUpdateFlags,
		orctinkgcpkms.M,
	)

	orc.Command(KeysCommand, orc.Modules(
		generateFlags,
	), cobra.Command{
		Use:   "generate",
		Short: "Generate a new set of keys",
	}, func() error {
		if keyOwnerFlag == "" {
			return fmt.Errorf("missing --canonical_host (canonical host, or other alias, of key owner)")
		}

		keys, err := orckeys.Generate(keyOwnerFlag)
		if err != nil {
			return err
		}

		if err := maybeUpdateRegistry(keys.Public()); err != nil {
			return err
		}

		return keys.WriteEncrypted(os.Stdout, masterKeyURI)
	})

	orc.Command(KeysCommand, orc.Modules(
		persistentkeys.M,
		publickeyregistry.M,
		recipientFlags,
	), cobra.Command{
		Use:   "sign-and-encrypt",
		Short: "Sign and encrypt data",
	}, func() error {
		var data interface{}

		allData, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(allData, &data); err != nil {
			return err
		}

		ciphertext, err := cryptopacket.Pack(data, persistentkeys.M.Keys, publickeyregistry.M, recipientFlag)
		if err != nil {
			return err
		}

		fmt.Println(ciphertext)
		return nil
	})

	orc.Command(KeysCommand, orc.Modules(
		persistentkeys.M,
		publickeyregistry.M,
	), cobra.Command{
		Use:   "verify-and-decrypt",
		Short: "Verify and decrypt data",
	}, func() error {
		ciphertext, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		packet, err := cryptopacket.Unpack(nil, string(ciphertext), persistentkeys.M.Keys, publickeyregistry.M)
		if err != nil {
			return err
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(packet)
		return nil
	})

	orc.Command(KeysCommand, orc.Modules(
		generateFlags,
	), cobra.Command{
		Use:   "generate-k8s",
		Short: "Print shell commands to generate a Kubernetes secret with keys",
	}, func() error {
		if keyOwnerFlag == "" {
			return fmt.Errorf("missing --canonical_host (canonical host, or other alias, of key owner)")
		}

		index := -1
		for i, arg := range os.Args {
			if arg == "generate-k8s" {
				index = i
				break
			}
		}
		if index == -1 {
			return fmt.Errorf("Unable to find self in command line")
		}

		prefixArgs := os.Args[:index]
		suffixArgs := os.Args[index+1:]

		prefix := strings.Join(prefixArgs, " ")
		suffix := strings.Join(suffixArgs, " ")
		fmt.Printf("%s generate %s | kubectl create secret generic orckeys-%s --from-file=keys.json=/dev/stdin\n", prefix, suffix, keyOwnerFlag)

		return nil
	})
}
