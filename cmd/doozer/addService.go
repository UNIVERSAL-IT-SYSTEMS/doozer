package main

func init() {
	cmds["addService"] = cmd{addService, "<serviceName> <scriptPath> <logPath> <redo>", "Adds a service"}
	cmdHelp["addService"] = `Adds the service with name <serviceName>, <scriptPath>, <logPath>, <redo>

`	
}

func addService(serviceName string, scriptPath string, logPath string, redo string) {
	c := dial()

	rev, err := c.Rev()

	if err != nil {
		panic(err)
	}

	c.Set("/watchedService/" + serviceName, rev, 
		[]byte(serviceName + "," + scriptPath + "," + logPath + "," + redo))

	c.Set("/" + serviceName + "/watching", rev, []byte("true"))

}
