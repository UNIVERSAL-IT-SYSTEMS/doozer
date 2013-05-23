package main

func init() {
	cmds["delService"] = cmd{delService, "<serviceName>", "Deletes a service"}
	cmdHelp["delService"] = `Deletes the service with name <serviceName>.
`
}

func delService(serviceName string) {
	c := dial()

	rev, err := c.Rev()

	if err != nil {
		panic(err)
	}

	c.Set("/" + serviceName + "/watching", rev, []byte("false"))

	c.Del("/watchedService/" + serviceName, rev)

}