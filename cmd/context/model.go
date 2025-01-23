package context

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/BytemanD/go-console/console"
	"github.com/wxnacy/wgo/file"
	"gopkg.in/yaml.v3"
)

const DEFAULT_CONTEXT_FILE = ".skyman_context.yaml"

func getContextFilePath() (f string) {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(u.HomeDir, DEFAULT_CONTEXT_FILE)
}

type Context struct {
	Name string `yaml:"name"`
	Conf string `yaml:"conf"`
}

type ContextConf struct {
	Current  string     `yaml:"current"`
	Contexts []*Context `yaml:"contexts"`
	filePath string
	changed  bool
}

func (c *ContextConf) Save() (err error) {
	defer func() {
		if err == nil {
			c.changed = false
		}
	}()
	if !c.changed {
		return nil
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return
	}
	if c.filePath == "" {
		return fmt.Errorf("context file is empty")
	}
	console.Debug("save context to %s", c.filePath)
	if file.IsFile(c.filePath) {
		if f, err := os.OpenFile(c.filePath, os.O_CREATE, 0644); err == nil {
			f.Close()
		} else {
			panic(err)
		}
	}
	return os.WriteFile(c.filePath, data, 0644)
}

func (c *ContextConf) SetContext(name string, conf string) {
	for _, context := range c.Contexts {
		if context.Name != name {
			continue
		}
		if context.Conf == conf {
			return
		}
		context.Conf = conf
		c.changed = true
		return
	}
	c.Contexts = append(c.Contexts, &Context{Name: name, Conf: conf})
	c.changed = true
}
func (c *ContextConf) RemoveContext(name string) {
	index := -1
	for i, context := range c.Contexts {
		if context.Name == name {
			index = i
			break
		}
	}
	if index >= 0 {
		copy(c.Contexts[index:], c.Contexts[index+1:])
		c.Contexts = c.Contexts[:len(c.Contexts)-1]
		c.changed = true
	}
	if c.Current == name {
		c.Current = ""
		c.changed = true
	}
}

func (c *ContextConf) SetCurrent(name string) error {
	for _, ctx := range c.Contexts {
		if ctx.Name == name {
			c.Current = name
			c.changed = true
			return nil
		}
	}
	return fmt.Errorf("context %s not exists", name)
}
func (c *ContextConf) GetCurrent() *Context {
	if c.Current == "" {
		return nil
	}
	for _, ctx := range c.Contexts {
		if ctx.Name == c.Current {
			return ctx
		}
	}
	return nil
}
func (c *ContextConf) Reset() {
	if c.Current == "" {
		return
	}
	c.Current = ""
	c.changed = true
}

func UseCluster(name string) error {
	cConf, err := LoadContextConf()
	if err != nil {
		return fmt.Errorf("load context failed: %s", err)
	}
	if err := cConf.SetCurrent(name); err != nil {
		return err
	}
	return cConf.Save()
}
