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

type Op struct {
	Name string
	Desc string
}

var InitPrompt = color.RedString(" _____         _       _ _     _   \n") + color.GreenString("|_   _|__   __| | ___ | (_)___| |_ \n") + color.YellowString("  | |/ _ \\ / _` |/ _ \\| | / __| __|\n") + color.BlueString("  | | (_) | (_| | (_) | | \\__ \\ |_ \n") + color.MagentaString("  |_|\\___/ \\__,_|\\___/|_|_|___/\\__|\n                                   \n")

var (
	OpList   = fmt.Sprintf("%s   %s", color.New(color.FgGreen).Sprint("List"), "[list all items]")
	OpAdd    = fmt.Sprintf("%s    %s", color.New(color.FgGreen).Sprint("Add"), "[add an undone item]")
	OpDelete = fmt.Sprintf("%s %s", color.New(color.FgGreen).Sprint("Delete"), "[delete a done/undone item]")
	OpDone   = fmt.Sprintf("%s   %s", color.New(color.FgGreen).Sprint("Done"), "[marking an item done]")
	OpSearch = fmt.Sprintf("%s %s", color.New(color.FgGreen).Sprint("Search"), "[search items by keyword]")
)

var ErrNoMatchItems = errors.New("no macth items")

var ops = []Op{
	{"List", OpList},
	{"Add", OpAdd},
	{"Delete", OpDelete},
	{"Done", OpDone},
	{"Search", OpSearch},
}

var (
	position int
	//operations = []string{OpList, OpAdd, OpDelete, OpDone, OpSearch}
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

// 这里先初始化
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
		Short:  "Select an operation",
		RunE:   RunRoot,
		PreRun: PreRunRoot,
	}
}

// 这是在正式执行命令前做的事情
// 这里输出了一些格式化内容
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

// 执行的命令
func RunRoot(cmd *cobra.Command, args []string) (err error) {
	template := promptui.SelectTemplates{
		Label:    "{{ . |  green}}",
		Active:   "\U0001F31F {{ .Name | green }}",
		Inactive: "  {{ .Name | white }}",
		Selected: "\U0001F31F {{ .Name | green }}",
		Details: `
--------- Description ----------
{{ .Desc }}`,
	}
	// 输出所有的菜单
	prompt := promptui.Select{
		Label:     "Select an operation",
		Items:     ops,
		CursorPos: position,
		Templates: &template,
	}
	i, _, err := prompt.Run()
	if err != nil {
		return err
	}
	op := ops[i].Desc
	// 判断选择的菜单是什么来执行对应的操作
	switch op {
	case OpList:
		position = 0
		if list.GlobalLists.IsEmpty() {
			fmt.Println(color.RedString("No items"))
			return
		}
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
		if list.GlobalLists.IsEmpty() {
			EmptyPrompt()
			return
		}
		err = DeleteOperation()
		if err != nil {
			FailedPrompt()
			return
		}
		SuccessPrompt()
	case OpDone:
		position = 3
		if list.GlobalLists.IsEmpty() {
			EmptyPrompt()
			return
		}
		err = DoneOperation()
		if err != nil {
			if err == ErrNoMatchItems {
				NoMatchPrompt()
				return nil
			}
			FailedPrompt()
			return
		}
		SuccessPrompt()
	case OpSearch:
		position = 4
		err = SearchOperation()
		if err != nil {
			if err == ErrNoMatchItems {
				NoMatchPrompt()
				return nil
			}
			FailedPrompt()
		}
		return
	}
	return
}

func FailedPrompt() {
	color.New(color.FgRed).Add(color.Bold).Println("failed")
}

func SuccessPrompt() {
	color.New(color.FgGreen).Add(color.Bold).Println("success")
}

func EmptyPrompt() {
	fmt.Println(color.RedString("No items"))
}

func NoMatchPrompt() {
	fmt.Println(color.RedString("No match items"))
}

func ListOperation() {
	var output strings.Builder
	output.Reset()
	// 查询目前所有的todo项
	for _, i := range list.GlobalLists.List() {
		_, err := output.WriteString(i.Display() + "\n")
		if err != nil {
			panic(err)
		}
	}
	//输出
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
	// 添加一项
	list.GlobalLists.Add(v)
	if err != nil {
		return err
	}
	return nil
}

func DeleteOperation() error {
	// 先查询出所有的item
	// 然后执行删除
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
	if err != nil {
		return err
	}
	return nil
}

func DoneOperation() error {
	// 先查询出所有的未完成的item
	// 然后执行标记已完成
	items := list.GlobalLists.ListSliceUndone()
	if len(items) == 0 {
		return ErrNoMatchItems
	}
	prompt := promptui.Select{
		Label: "Select one undone item",
		Items: items,
	}
	i, _, err := prompt.Run()
	if err != nil {
		return err
	}
	if i < 0 {
		return errors.New("wrong index")
	}
	list.GlobalLists.Done(i)
	if err != nil {
		return err
	}
	return nil
}

func SearchOperation() error {
	prompt := promptui.Prompt{
		Label: "input keyword",
	}
	// 读取用户输入的关键字
	v, err := prompt.Run()
	if err != nil {
		return err
	}
	var output strings.Builder
	output.Reset()
	// 这里去list中根据关键字匹配
	items := list.GlobalLists.Search(v)
	if len(items) == 0 {
		return ErrNoMatchItems
	}
	for _, i := range items {
		_, err := output.WriteString(i.Display() + "\n")
		if err != nil {
			panic(err)
		}
	}
	// 将结果输出
	fmt.Println(output.String())
	return nil
}
