## envsec set

Securely store one or more environment variables

### Synopsis

Securely store one or more environment variables. To test contents of a file as a secret use set=@<file>

```
envsec set <NAME1>=<value1> [<NAME2>=<value2>]... [flags]
```

### Options

```
      --environment string   Environment name, such as dev or prod (default "dev")
  -h, --help                 help for set
      --org-id string        Organization id to namespace secrets by
      --project-id string    Project id to namespace secrets by
```

### SEE ALSO

* [envsec](envsec.md)	 - Manage environment variables and secrets

