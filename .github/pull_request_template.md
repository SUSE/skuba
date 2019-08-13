## Why is this PR needed?

Does it fix an issue? addresses a business case?

add a description and link to the issue if one exists.

Fixes #

**Reminder**: Add the "fixes bsc#XXXX" to the title of the commit so that it will
appear in the changelog.


## What does this PR do?

please include a brief "management" technical overview (details are in the code)

## Anything else a reviewer needs to know?

Special test cases, manual steps, links to resources or anything else that could be helpful to the reviewer.

## Info for QA

This is info for QA so that they can validate this. This is **mandatory** if this PR fixes a bug.
If this is a new feature, a good description in "What does this PR do" may be enough.

### Related info

Info that can be relevant for QA:
* link to other PRs that should be merged together
* link to packages that should be released together
* upstream issues

### Status **BEFORE** applying the patch

How can we reproduce the issue? How can we see this issue? Please provide the steps and the prove
this issue is not fixed.

### Status **AFTER** applying the patch

How can we validate this issue is fixed? Please provide the steps and the prove this issue is fixed.


## Docs

If docs need to be updated, please add a link to a PR to https://github.com/SUSE/doc-caasp.
At the time of creating the issue, this PR can be work in progress (set its title to [WIP]),
but the documentation needs to be finalized before the PR can be merged. 


# Merge restrictions after RC

(Please do not edit this)

This is the "better safe than sorry" phase, so we will restrict who and what can be merged to prevent unexpected surprises:

    Who can merge: Release Engineers*
    What can be merged (merge criteria):
        3 approvals:
            1 developer
            1 manager (Nuno, Flavio, Klaus, Markos, Stef, David, Rafa or Jordi)
            1 QA
        there is a PR for updating documentation (or a statement that this is not needed)
        it fixes a P2 bug that has not been target for maintenance
        the impact is low: for example, it does not touch a piece of code that has an impact on all features

* *Managers will keep the permissions, but they are supposed to give approval. Merging will be done by Release Engineers because they will coordinate the release of the fix.*

<!-- Remember, if this is a work in progress please pre-append [WIP] to the title until you are ready! 
    If you can, please apply all applicable labels to help reviews out! -->
