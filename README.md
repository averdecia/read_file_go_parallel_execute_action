# script_command
The commands will read files and execute some actions with it. The command will handle the opening files, saving error, report progress and velocity, closing cursor

To use the lib you should create a class that implement the ICommand interface.

```go
type MyCommand struct {}

func (c *MyCommand) ExecuteAction(element []string) (string, error) {
   fmt.Printf("Element: %v", element)
   return "OK", nil
}
```

Then you must call the RunProcess function passing the command, input and output filepaths, and the amount of routines to run in parallel

```go
func main(){
   mycommand := &MyCommand{}
   command.RunProcess(mycommand, 10, "/tmp/exampleInput.csv", "/tmp/errorFiles.csv")      // You can use os.Args[1:] to pass the variables to the build 
}
```
