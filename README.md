# gsync
Nerfed rsync

```
Usage of gsync:
gsync [options] sourceDir destDir
  -v verbose
  -f try to chmod unwriteable destination
  -m XYZ write files/directories with octal mode XYZ
```

gsync recursively copies sourceDir to destDir,
creating directories as necessary in the destination.
By default the file modes are copied; -m overrides this.
If the destination files are not writeable, -f will
attempt to make them writeable (and then restore to
the desired protection afterwards).
