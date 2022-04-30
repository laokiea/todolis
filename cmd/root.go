package cmd

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/laokiea/todolist/list"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var InitPrompt = color.RedString(" _____         _       _ _     _   \n") + color.GreenString("|_   _|__   __| | ___ | (_)___| |_ \n") + color.YellowString("  | |/ _ \\ / _` |/ _ \\| | / __| __|\n") + color.BlueString("  | | (_) | (_| | (_) | | \\__ \\ |_ \n") + color.MagentaString("  |_|\\___/ \\__,_|\\___/|_|_|___/\\__|\n                                   \n")

var (
	OpList   = fmt.Sprintf("%s   %s", color.New(color.FgGreen).Sprint("List"), "[list all items]")
	OpAdd    = fmt.Sprintf("%s    %s", color.New(color.FgGreen).Sprint("Add"), "[add an undone item]")
	OpDelete = fmt.Sprintf("%s %s", color.New(color.FgGreen).Sprint("Delete"), "[delete a done/undone item]")
	OpDone   = fmt.Sprintf("%s   %s", color.New(color.FgGreen).Sprint("Done"), "[marking an item done]")
	OpSearch = fmt.Sprintf("%s %s", color.New(color.FgGreen).Sprint("Search"), "[search items by keyword]")
)

var (
	position   int
	operations = []string{OpList, OpAdd, OpDelete, OpDone, OpSearch}
)

var (
	o                 sync.Once
	initPromptDisplay bool
	DefaultCommand    *cobra.Command
)

func Execute() error {
	err := DefaultCommand.Execute()
	return err
}

func NewCommand() *cobra.Command {
	o.Do(func() {
		DefaultCommand = &cobra.Command{
			Use:    "Todolist",
			Short:  "Make up your own Todolist component",
			Long:   "A little artful budget by Miss SicilySunset",
			RunE:   RunRoot,
			PreRun: PreRunRoot,
		}
	})
	return DefaultCommand
}

func DefaultCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "default",
		Short:  "Select a operation",
		RunE:   RunRoot,
		PreRun: PreRunRoot,
	}
}

func PreRunRoot(cmd *cobra.Command, args []string) {
	if !initPromptDisplay {
		fmt.Println(InitPrompt)
		err := list.GlobalLists.Load()
		if err != nil {
			panic(err)
		}
		initPromptDisplay = true
	}
}

func RunRoot(cmd *cobra.Command, args []string) (err error) {
	prompt := promptui.Select{
		Label:     "Select an operation",
		Items:     operations,
		CursorPos: position,
	}
	_, op, err := prompt.Run()
	if err != nil {
		return err
	}
	switch op {
	case OpList:
		position = 0
		ListOperation()
	case OpAdd:
		position = 1
		err = AddOperation()
		if err != nil {
			FailedPrompt()
			return
		}
		SuccessPrompt()
	case OpDelete:
		position = 2
		err = DeleteOperation()
		if err != nil {
			FailedPrompt()
			return
		}
		SuccessPrompt()
	case OpDone:
		position = 3
		err = DoneOperation()
		if err != nil {
			FailedPrompt()
			return
		}
	case OpSearch:
		position = 4
		SearchOperation()
	}
	return
}

func FailedPrompt() {
	color.New(color.FgRed).Add(color.Bold).Println("failed")
}

func SuccessPrompt() {
	color.New(color.FgGreen).Add(color.Bold).Println("success")
}

func ListOperation() {
	var output strings.Builder
	output.Reset()
	for _, i := range list.GlobalLists.List() {
		_, err := output.WriteString(i.Display() + "\n")
		if err != nil {
			panic(err)
		}
	}
	fmt.Println(output.String())
}

func AddOperation() error {
	prompt := promptui.Prompt{
		Label: "input an item",
	}
	v, err := prompt.Run()
	if err != nil {
		return err
	}
	list.GlobalLists.Add(v)
	err = list.GlobalLists.Flush()
	if err != nil {
		return err
	}
	return nil
}

func DeleteOperation() error {
	prompt := promptui.Select{
		Label: "Select one item",
		Items: list.GlobalLists.ListSliceAll(),
	}
	i, _, err := prompt.Run()
	if err != nil {
		return err
	}
	if i < 0 {
		return errors.New("wrong index")
	}
	list.GlobalLists.Del(i)
	err = list.GlobalLists.Flush()
	if err != nil {
		return err
	}
	return nil
}

func DoneOperation() error {
	prompt := promptui.Select{
		Label: "Select one undone item",
		Items: list.GlobalLists.ListSliceUndone(),
	}
	i, _, err := prompt.Run()
	if err != nil {
		return err
	}
	if i < 0 {
		return errors.New("wrong index")
	}
	list.GlobalLists.Done(i)
	err = list.GlobalLists.Flush()
	if err != nil {
		return err
	}
	return nil
}

func SearchOperation() error {
	prompt := promptui.Prompt{
		Label: "input keyword",
	}
	v, err := prompt.Run()
	if err != nil {
		return err
	}
	var output strings.Builder
	output.Reset()
	for _, i := range list.GlobalLists.Search(v) {
		_, err := output.WriteString(i.Display() + "\n")
		if err != nil {
			panic(err)
		}
	}
	fmt.Println(output.String())
	return nil
}
