# tpkl

[Pkl][pkl] ♥️ tasks! Add Pkl to your task books!

```
$ cat tasks.pkl
import "tpkl:tpkl"
subject = tpkl.argv.getOrNull(0) ?? "world"
tasks: tpkl.Tasks = new {
  ["hello"] {
    cmds {
      "echo Hello, \(subject)!" |> tpkl.sh
    }
  }
}
$ tpkl hello
Hello, world!
$ tpkl hello Pkl
Hello, Pkl!
```

[pkl]: https://pkl-lang.org/
