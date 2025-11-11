# How to contribute

First, thanks for taking the time to contribute to our project! There are many ways you can help out.

### Questions

If you have a question that needs an answer, [create an issue](https://help.github.com/articles/creating-an-issue/), and
label it as a question.

### Issues for bugs or feature requests

If you encounter any bugs in the code, or want to request a new feature or enhancement,
please [create an issue](https://help.github.com/articles/creating-an-issue/) to report it. Kindly add a label to
indicate what type of issue it is.

### Run Locally

A minikube lab exists to run a few dummy applications in different language to easily tests the profilers.

Build all components and start the cluster with:

```
make minikube-all
```

You can build the plugin with

```
make build-cli
```

You can then profile any application by adding the plugin to your path and call minikube kubectl:

```
export PATH=./bin/:$PATH 
minikube kubectl -- prof \
    -n stupid-apps $POD \
    --image localhost/docker/jvm:latest --image-pull-policy Never \
    -t 10s -e wall -l java --tool async-profiler 
```

 * Add the plugin built from source to your `PATH`. 
 * Dummy apps runs in the `stupid-apps` namespace.
 * To use an image built from source, use the `--image` option and `--image-pull-policy Never`.

### Contribute Code

We welcome your pull requests for bug fixes. To implement something new, please create an issue first so we can discuss
it together.

***Creating a Pull Request***
Please follow [best practices](https://github.com/trein/dev-best-practices/wiki/Git-Commit-Best-Practices) for creating
git commits.

When your code is ready to be
submitted, [submit a pull request](https://help.github.com/articles/creating-a-pull-request/) to begin the code review
process.

We only seek to accept code that you are authorized to contribute to the project. We have added a pull request template
on our projects so that your contributions are made with the following confirmation:

> I confirm that this contribution is made under the terms of the license found in the root directory of this repository's source tree and that I have the authority necessary to make this contribution on behalf of its copyright owner.
