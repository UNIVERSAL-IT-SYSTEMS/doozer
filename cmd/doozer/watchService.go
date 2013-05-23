package main

import (
	"fmt"
	"strings"
	"os"
	"github.com/ha/doozer"
	"syscall"
)

var watchers map[string]chan bool

func init() {
	cmds["watchService"] = cmd{watchService, "<reWatch> <redoScript>", "Begin to watch services"}
	cmdHelp["watchService"] = ` set <reWatch>=true to watch the existing services 
	set <redoScript>=true to redo the existing key value pairs

`
}

func watchService(reWatch string, redoScript string) {
	c := dial()

	if *rrev == -1 {
		var err error
		*rrev, err = c.Rev()
		if err != nil {
			bail(err)
		}
	}

	watchServicePath := "/watchedService" 

	// init map
	watchers = make(map[string]chan bool)

	//init work, add the existing watch service
	if reWatch == "true" {

		v := make(vis)
		errs := make(chan error)

		go func() {
			doozer.Walk(c, *rrev, watchServicePath, v, errs)
			close(v)
			}()

			var finish bool 
			finish = false

			for {
				if finish {
					break
				}
				select {
				case path, ok := <-v:
					if !ok {
						finish = true
						break
					}
					if path != watchServicePath {

						body, _, err := c.Get(path, nil)
						if err != nil {
							bail(err)
						}
						register(c, body, redoScript)
					}

				case err := <-errs:
					fmt.Fprintln(os.Stderr, err)
				}
			}
	}

	// start to watch
	watchServicePath += "/**"
	for {

		// func Wait would block...
		ev, err := c.Wait(watchServicePath, *rrev)
		if err != nil {
			bail(err)
		}
		*rrev = ev.Rev + 1

		switch {
		case ev.IsSet():
			register(c, ev.Body, redoScript)
		case ev.IsDel():
			continue
		}
	}
}

func register(c *doozer.Conn, body []byte, redoScript string) {

	entries := strings.Split(string(body), ",")
	watchPath := "/" + entries[0]
	scriptPath := entries[1]
	filePath := entries[2]
	redo := entries[3]

	if redo == "false" {
		redoScript = "false"
	}

	// create a channels for the watcher and the main routines
	// to communicate
	ch := make(chan bool)

	// add the new watcher into watcher map
	watchers[watchPath] = ch


	setWatchPath := watchPath + "/watching"
	rev, err := c.Rev()
	if err != nil {
		bail(err)
	}
	c.Set(setWatchPath, rev, []byte("true"))

	// create new etcd Goroutines
	go watcher(watchPath, filePath, scriptPath, redoScript, ch)

	fmt.Println("Registered a watch service\nLog:", 
		filePath, "\nScript:", scriptPath, "\nWatchedPath:", 
		watchPath)	
	fmt.Println("--------------------------------")

}

func watcher(glob string, filePath string, scriptPath string, redo string, stop chan bool) {
	c := dial()

	if *rrev == -1 {
		var err error
		*rrev, err = c.Rev()
		if err != nil {
			bail(err)
		}
	}

	// open the file
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND, 0660) // For append.

	if err != nil {
		fmt.Println("open file error")
		panic(err)
	}

	// redo previous works
	if redo == "true" {
		v := make(vis)
		errs := make(chan error)

		go func() {
			doozer.Walk(c, *rrev, glob, v, errs)
			close(v)
			}()

		finish := false
		for {
			select {
			case path, ok := <-v:
				if !ok {
					finish = true
					break
				}
				if path != glob {
					body, _, err := c.Get(path, nil)
					if err != nil {
						bail(err)
					}
					if _, err := file.WriteString("REDO.......\n" + path + " "); 
					err != nil {
						panic(err)
					}

					if _, err := file.Write(body); err != nil {
						panic(err)
					}

					if _, err := file.WriteString("\n"); err != nil {
						panic(err)
					}

					fmt.Println()

					if path == glob + "/watching"  {
						if string(body) == "false" {
							fmt.Println("exiting")
							return
						} else {
							continue
						}

					} 

					runScript(scriptPath, path, body)
				}

			case err := <-errs:
				fmt.Fprintln(os.Stderr, err)
			}
			if finish {
				break
			}
		}
	}

	watchPath := glob + "/**"
	// real work begin here
	for {
		
		// func Wait would block...
		ev, err := c.Wait(watchPath, *rrev)
		if err != nil {
			bail(err)
		}
		*rrev = ev.Rev + 1

		switch {
		case ev.IsSet():

			// write a change to the log file
			// seems that golang directly call write(), no need to flush
			// bad performance

			if _, err := file.WriteString(ev.Path + " "); err != nil {
           		panic(err)
        	}

        	if _, err := file.Write(ev.Body); err != nil {
            	panic(err)
        	}

        	if _, err := file.WriteString("\n"); err != nil {
            	panic(err)
        	}

        	if ev.Path == glob + "/watching"  {
        		if string(ev.Body) == "false" {
        			fmt.Println("exiting")
        			return
        		} else {
        			continue
        		}
        		
        	} 

        	runScript(scriptPath, ev.Path, ev.Body)
		case ev.IsDel():
			continue
		}

	}
}

func runScript(scriptPath string, key string, value []byte) {
	// fork and exec the given script
	var procAttr syscall.ProcAttr

	// is new line char at the end?
	if value[len(value) - 1] == 10 {
		value = value[:len(value) - 1]
	} 	

	argv := []string{scriptPath, key, string(value)} 
	if _, err := syscall.ForkExec(scriptPath, argv, &procAttr); err != nil {
		fmt.Println("exec error")
		panic(err)
	}
	fmt.Print("runScript[" + scriptPath + "]  withArgs")
	fmt.Println(argv)
	fmt.Println("--------------------------------")
}