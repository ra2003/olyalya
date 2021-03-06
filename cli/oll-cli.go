package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/dzyanis/olyalya/pkg/client"
	"github.com/dzyanis/olyalya/pkg/cmd"
	"os"
	"strconv"
	"strings"
)

var (
	ErrNotEnoughArguments = errors.New("Not enough arguments")
	ErrCanNotExit         = errors.New("Something really went wrong")
)

var (
	httpAddr = flag.String("http.addr", "http://localhost:3000", "HTTP listen URL")
	version  = flag.Bool("version", false, "Print a version")
)

const (
	HelpInformation = `Command is not exist.
		Run 'HELP' for usage or read more on https://github.com/dzyanis/olyalya
	`
)

var (
	buildName = ""
	Client    *client.Client
	Cmd       = cmd.NewCmd()
)

func main() {
	if *version {
		fmt.Printf("O(lya-lya) client build: %s\n", buildName)
		os.Exit(0)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("O(lya-lya) greets you")

	for {
		fmt.Printf("%s> ", Client.CurrentInstanceName())
		cli, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			continue
		}

		result, err := Cmd.Run(cli)
		if err != nil {
			if cmd.ErrCommandNotExist == err {
				fmt.Printf(HelpInformation)
			} else {
				fmt.Printf("ERROR: %s\n", err.Error())
			}
			continue
		}
		fmt.Println(result)
	}
}

func ValidCountArguments(args []string, min int) error {
	if len(args) < min {
		return ErrNotEnoughArguments
	}
	return nil
}

func ValidInt(s string) (int, error) {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return i, err
}

func ValidName(s string) (string, error) {
	return s, nil
}

func ValidString(s string) (string, error) {
	res := strings.Trim(s, `"`)
	return res, nil
}

func ValidArray(s string) ([]string, error) {
	var arr []string
	err := json.Unmarshal([]byte(s), &arr)
	return arr, err
}

func ValidHash(s string) (map[string]string, error) {
	var h map[string]string
	err := json.Unmarshal([]byte(s), &h)
	return h, err
}

func handlerHelp(c *cmd.Cmd, args []string, line string) (string, error) {
	if len(args) > 1 {
		command, ok := c.Commands[args[1]]
		if !ok {
			return "", cmd.ErrCommandNotExist
		}

		return fmt.Sprintf("%s\n%s", command.Title, command.Description), nil
	}

	command, _ := c.Commands[args[0]]
	result := fmt.Sprintf("%s\n%s\n\nList of commands:", command.Title, command.Description)
	for name, command := range c.Commands {
		result = result + fmt.Sprintf("\n%s - %s", name, command.Title)
	}

	return result, nil
}

func handlerList(c *cmd.Cmd, args []string, line string) (string, error) {
	list, err := Client.ListInstances()
	if err != nil {
		return "", err
	}
	result := ""
	for i, e := range list {
		result = result + fmt.Sprintf("%d) %s\n", i+1, e)
	}
	return strings.Trim(result, "\n"), nil
}

func handlerSetTTL(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 3)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	ttl, err := ValidInt(args[2])
	if err != nil {
		return "", err
	}

	err = Client.SetTTL(name, ttl)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func handlerDeleteTTL(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 2)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	err = Client.DelTTL(name)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func handleArrayElementGet(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 3)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	index, err := ValidInt(args[2])
	if err != nil {
		return "", err
	}

	return Client.GetArrayElement(name, index)
}

func handleEcho(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 2)
	if err != nil {
		return "", err
	}

	return ValidString(args[1])
}

func handleExit(c *cmd.Cmd, args []string, line string) (string, error) {
	fmt.Println("Bye!")
	os.Exit(0)
	return "", ErrCanNotExit
}

func handleInstanceCreate(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 2)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	err = Client.CreateInstance(name)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func handleInstanceSelect(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 2)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	err = Client.SelectInstance(name)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func handleKeys(c *cmd.Cmd, args []string, line string) (string, error) {
	list, err := Client.Keys()
	if err != nil {
		return "", err
	}

	result := ""
	for ind, key := range list {
		result = result + fmt.Sprintf("%d) %s\n", ind+1, key)
	}

	return result, nil
}

func handleInstanceSet(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 3)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	str, err := ValidString(args[2])
	if err != nil {
		return "", err
	}

	ttl := 0
	if len(args) > 3 {
		ttl, _ = strconv.Atoi(args[3])
	}

	err = Client.Set(name, str, ttl)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func handlerHashSet(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 3)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	h, err := ValidHash(args[2])
	if err != nil {
		return "", err
	}

	ttl := 0
	if len(args) > 3 {
		ttl, _ = strconv.Atoi(args[3])
	}

	err = Client.SetHash(name, h, ttl)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func handleInstanceArraySet(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 3)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	fmt.Println(args[2])
	arr, err := ValidArray(args[2])
	if err != nil {
		fmt.Println(arr, err)
		return "", err
	}
	fmt.Println(arr, err)

	ttl := 0
	if len(args) > 3 {
		ttl, _ = strconv.Atoi(args[3])
	}

	err = Client.SetArray(name, arr, ttl)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func handleArrayGet(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 2)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	a, err := Client.GetArray(name)
	return fmt.Sprintf("%v", a), err
}

func handleHashGet(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 2)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	h, err := Client.GetHash(name)
	return fmt.Sprintf("%v", h), err
}

func handleInstanceDel(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 2)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	err = Client.Del(name)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func handleInstanceGet(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 2)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	value, err := Client.Get(name)
	if err != nil {
		return "", err
	}

	return value, nil
}

func handleArrayElementAdd(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 2)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	s, err := ValidString(args[2])
	if err != nil {
		return "", err
	}

	err = Client.AddArrayElement(name, s)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func handleArrayElementSet(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 4)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	index, err := ValidInt(args[2])
	if err != nil {
		return "", err
	}

	value, err := ValidString(args[3])
	if err != nil {
		return "", err
	}

	err = Client.SetArrayElement(name, index, value)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func handleArrayElementDel(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 3)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	index, err := ValidInt(args[2])
	if err != nil {
		return "", err
	}

	err = Client.DelArrayElement(name, index)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func handleHashElementGet(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 3)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	key, err := ValidName(args[2])
	if err != nil {
		return "", err
	}

	res, err := Client.GetHashElement(name, key)
	if err != nil {
		return "", err
	}

	return res, nil
}

func handleHashElementSet(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 4)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	key, err := ValidName(args[2])
	if err != nil {
		return "", err
	}

	val, err := ValidString(args[3])
	if err != nil {
		return "", err
	}

	err = Client.SetHashElement(name, key, val)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func handleHashElementDel(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 3)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	key, err := ValidName(args[2])
	if err != nil {
		return "", err
	}

	err = Client.DelHashElement(name, key)
	if err != nil {
		return "", err
	}

	return "OK", nil
}
func handleDestroy(c *cmd.Cmd, args []string, line string) (string, error) {
	err := ValidCountArguments(args, 2)
	if err != nil {
		return "", err
	}

	name, err := ValidName(args[1])
	if err != nil {
		return "", err
	}

	err = Client.Destroy(name)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

var (
	commands = map[string]*cmd.Command{
		"HELP": {
			Title:       "Function show information about other functions",
			Description: "Example: HELP <FUNCTION_NAME>",
			Handler:     handlerHelp,
		},
		"ECHO": {
			Title:       "Prints string",
			Description: "Example: ECHO \"Hello World!\"",
			Handler:     handleEcho,
		},
		"EXIT": {
			Title:       "Exit from the program",
			Description: "",
			Handler:     handleExit,
		},

		"CREATE": {
			Title:       "Create an instance",
			Description: "Example: CREATE dbname",
			Handler:     handleInstanceCreate,
		},

		"LIST": {
			Title:       "Show list of instance",
			Description: "Example: LIST",
			Handler:     handlerList,
		},

		"SELECT": {
			Title:       "Select an instance",
			Description: "Example: SELECT instance_name",
			Handler:     handleInstanceSelect,
		},

		"KEYS": {
			Title:       "Show all keys",
			Description: "Example: KEYS",
			Handler:     handleKeys,
		},
		"SET": {
			Title:       "Set value",
			Description: "Example: SET name \"value\" ttl",
			Handler:     handleInstanceSet,
		},
		"GET": {
			Title:       "Get value",
			Description: "Example: GET name",
			Handler:     handleInstanceGet,
		},
		"DEL": {
			Title:       "Delete value",
			Description: "Example: DEL name",
			Handler:     handleInstanceDel,
		},
		"TTL/SET": {
			Title:       "Set time to live",
			Description: "Example: TTL/SET mayfly 86400",
			Handler:     handlerSetTTL,
		},
		"TTL/DEL": {
			Title:       "Remove time to live",
			Description: "Example: TTL/DEL mayfly",
			Handler:     handlerDeleteTTL,
		},
		"ARR/SET": {
			Title:       "Set array",
			Description: "Example: ARR/SET name [] ttl",
			Handler:     handleInstanceArraySet,
		},
		"ARR/GET": {
			Title:       "Get array",
			Description: "Example: ARR/GET name",
			Handler:     handleArrayGet,
		},
		"ARR/EL/GET": {
			Title:       "Returns the element associated with index",
			Description: "Example: ARR/EL/GET name index",
			Handler:     handleArrayElementGet,
		},
		"ARR/EL/ADD": {
			Title:       "Add the element to an array",
			Description: "Example: ",
			Handler:     handleArrayElementAdd,
		},
		"ARR/EL/SET": {
			Title:       "Set the element of an array",
			Description: "Example: ",
			Handler:     handleArrayElementSet,
		},
		"ARR/EL/DEL": {
			Title:       "Delete the element of an array",
			Description: "Example:",
			Handler:     handleArrayElementDel,
		},
		"HASH/GET": {
			Title:       "Get a hash",
			Description: "Example: HASH/GET name",
			Handler:     handleHashGet,
		},
		"HASH/SET": {
			Title:       "Set a hash",
			Description: "Example: HASH/SET name {}",
			Handler:     handlerHashSet,
		},
		"HASH/EL/GET": {
			Title:       "Get the element of a hash",
			Description: "Example:",
			Handler:     handleHashElementGet,
		},
		"HASH/EL/SET": {
			Title:       "Set the element of a hash",
			Description: "Example:",
			Handler:     handleHashElementSet,
		},
		"HASH/EL/DEL": {
			Title:       "Delete the element of a hash",
			Description: "Example:",
			Handler:     handleHashElementDel,
		},
		"DESTROY": {
			Title:       "Remove instance",
			Description: "Example: DESTROY instance_name",
			Handler:     handleDestroy,
		},
	}
)

func init() {
	flag.Parse()

	Client = client.NewClient(*httpAddr)

	for name, command := range commands {
		Cmd.Add(name, command)
	}
}
