# Contributing Guidelines

Thank you for your interest in contributing to our project. Whether it's a bug report, new feature, correction, or additional
documentation, we greatly value feedback and contributions from our community.

Please read through this document before submitting any issues or pull requests to ensure we have all the necessary
information to effectively respond to your bug report or contribution.


## Security issue notifications
If you discover a potential security issue in this project we ask that you notify AWS/Amazon Security via our [vulnerability reporting page](http://aws.amazon.com/security/vulnerability-reporting/). Please do **not** create a public github issue.


## Reporting Bugs/Feature Requests

We welcome you to use the GitHub issue tracker to report bugs or suggest features.

When filing an issue, please check existing open or recently closed issues to make sure somebody else hasn't already
reported the issue. Please include as much information as you can, such as:

* What you encountered and what you expected to occur instead
* A reproducible test case or series of steps
* The version of our code being used (ex: commit or tag link)
* Any modifications you've made relevant to the bug
* Which operating system(s) are you running on and which are you building for when encountering the bug
* Additional information about your environment or deployment that could be relevant such as:
  * Game Engines or external packages that could have dependency conflicts

Stay involved with your issue after creation. The GitHub label, requesting info, is applied by our engineers when we expect more information before proceeding.


## Contributing via Pull Requests
Contributions via pull requests are much appreciated. Before sending us a pull request, please ensure that:

1. You are working against the latest source on the *develop* branch. The *main* branch is exclusively for verified releases.
2. You check existing open, and recently merged, pull requests to make sure someone else hasn't addressed the problem already.
3. You open or respond to an existing issue to discuss plans for significant changes - we would hate for your time to be wasted if the work is already in progress or does not align with our plans.

To send us a pull request, please:

1. Fork the repository.
2. Please focus on the specific change you are contributing. If you also reformat all the code for example, it will be hard for us to review the main purpose of your change.
3. Add new tests to verify new logic.
4. Ensure local tests pass.
5. Commit to your fork using clear commit messages. See section below for a brief guide.
6. Send us a pull request, answering any default questions in the pull request interface.
7. Pay attention to any automated CI failures reported in the pull request and stay involved in the conversation.

GitHub provides additional document on [forking a repository](https://help.github.com/articles/fork-a-repo/) and
[creating a pull request](https://help.github.com/articles/creating-a-pull-request/).


## Finding contributions to work on
Most open issues are available for someone to work on. Please send a message in the conversation if you plan to work on a task.
This gives us a chance to alert you if it's already in progress and makes sure another contributor doesn't work on the same item.

Issues marked with the label "backlog" are being tracked by our engineers in our own backlog; however, it doesn't mean we've started working on it.


## Code of Conduct
This project has adopted the [Amazon Open Source Code of Conduct](https://aws.github.io/code-of-conduct).
For more information see the [Code of Conduct FAQ](https://aws.github.io/code-of-conduct-faq) or contact
opensource-codeofconduct@amazon.com with any additional questions or comments.


## Commit message guide

```
[#<issue>] Succinct summary of changes (aim for 50-72 character maximum)

If the subject is not enough, can include more detail below. Wrap at 72
characters with a blank line separating from the subject.

Paragraphs within the body should also be separated with a blank line.
```

Guidelines:
* The subject should be specific. Do not rely on the body to clarify context. Ex: "Fix UI crash when auth token is not populated" is preferred over "Fix ui bug".
* Write the subject in the imperative: "Fix bug" and not "Fixed bug" or "Fixes bug." This is to be consistent with git merge and git revert.
* Subject should reference the issue number being addressed. Even when the work spans multiple commits.
* If there is important context that isn't immediately apparent. The info should be in code comments, not in the commit body.


## Licensing

See the [LICENSE](LICENSE) file for our project's licensing. We will ask you to confirm the licensing of your contribution.

