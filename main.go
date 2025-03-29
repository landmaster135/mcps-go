package main

import (
	mypkg "example.com/mymodule/mypkg"
)

func main() {
	// env_name := "ANY_TOKEN"
	// TOKEN, ok := os.LookupEnv(env_name)
	// if !ok {
	// 	fmt.Printf("%s is not set", env_name)
	// }
	// results, err := mypkg.AnyFunction(TOKEN)
	// if err != nil {
	// 	mypkg.OutLog(err)
	// 	panic(err)
	// }
	mypkg.OutLog("main: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	// mypkg.OutLog(results)
	// j, err := json.Marshal(results)
	// b := bytes.NewBuffer([]byte(j))
	// mypkg.OutLog(b)

	// mypkg.OutLog(results)
	l := mypkg.NewBuiltinLogger("FILE")
	l.Debug("this is debug message")
	l.Info("this is info message")
	l.Warning("this is warning message")
	l.Error("this is error message")
	l.Fatal("this is fatal message")
}
