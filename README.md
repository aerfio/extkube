# extkube

Usage:
```bash
extkube --context ${K8S_CONTEXT}
```

This command will print:
```
Copied:
export KUBECONFIG="/var/folders/yz/7drt587x27g8cltqm8f1qfn80000gn/T/kubeconfig-2314541764"
to clipboard
```

and then you can paste that line into any terminal window, which will change KUBECONFIG only for this particular terminal window.

What happened under the hood:
1. A temporary file was created
2. kubeconfig with only 1 kube context was extracted from the active kubeconfig, usually it's `~/.kube/config`. `kubectl` is **not** needed to be in `$PATH`
3. That extracted kubeconfig is written to this temporary file
4. The string in following form:
    ```bash
    export KUBECONFIG=$PATH_TO_THAT_TEMPORARY_FILE
    ```
    is copied into your clipboard.

You can just paste that command into your terminal to get "scoped" kube context. I used it for example when developing multi-cluster Kubernetes Operators.

## Rust version

[extkube-rs](https://github.com/aerfio/extkube/tree/extkube-rs) git branch of this repository holds Rust implementation of this CLI tool, created for learning purposes.
