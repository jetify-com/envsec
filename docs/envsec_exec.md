## envsec exec

Execute a command with Jetify-stored environment variables

### Synopsis

Execute a specified command with remote environment variables being present for the duration of the command. If an environment variable exists both locally and in remote storage, the remotely stored one is prioritized.

```
envsec exec <command> [flags]
```

### Options

```
      --environment string   Environment name, such as dev or prod (default "dev")
  -h, --help                 help for exec
      --org-id string        Organization id to namespace secrets by
      --project-id string    Project id to namespace secrets by
```

### SEE ALSO

* [envsec](envsec.md)	 - Manage environment variables and secrets

