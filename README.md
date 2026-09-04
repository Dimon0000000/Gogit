# Gogit

> Git is easy.
>
> Until your mentor asks you to do something with it.

Imagine this.

You just started your first internship as a developer.

Day one:

```bash
git clone ...
```

Easy.

Day two:

```bash
git checkout -b feat/something
```

Still easy.

Then one day, your mentor walks over and says:

> "Main has moved forward. Rebase your commits onto the latest main,
> resolve the conflicts, continue the rebase, and update your remote branch safely."

You:

> "Sure."

Your brain:

> **What the hell is a rebase?**

So you start searching.

Google.

Stack Overflow.

GitHub.

AI.

`1` minutes later, you somehow assemble this:

```bash
git fetch origin
git rebase origin/main

# resolve conflicts...

git add .
git rebase --continue
git push --force-with-lease
```

It works.

For a brief moment, you believe you understand Git.

Then the next day your mentor says:

> "Actually, remove yesterday's commit, but keep the changes in your working tree."

You stare at the terminal.

```bash
git reset ???
```

`--soft`?

`--mixed`?

`--hard`?

Three minutes ago you were a software engineer.

Now you're Googling:

> **difference between git reset soft mixed hard**

---

This was basically my experience while interning on
**infiniflow/ragflow**.

Git doesn't have a shortage of commands.

If anything, it has **way too many of them**.

The problem is usually not:

> "I don't know what I want to do."

The problem is:

> "I know exactly what I want to do,
> but I have absolutely no idea what that damn flag is called."

So I built **Gogit**.

---

## What is Gogit?

**Gogit** is a Git CLI assistant written in Go.

It does not try to replace Git.

It simply tries to help when you're staring at this:

```bash
git branch --sho|
```

and wondering what comes next.

Gogit can suggest:

```text
--show-current
    Print the name of the current branch.
```

Press `Enter`:

```bash
git branch --show-current
```

Done.

No need to:

1. open a browser
2. search the Git documentation
3. open Stack Overflow
4. ask an AI
5. copy an answer written in 2014
6. pray it doesn't delete your working tree

The goal is to turn this:

```text
I know what I want to do
          ↓
       git ...
          ↓
    Gogit helps
          ↓
         Done
```

instead of this:

```text
I know what I want to do
          ↓
       Google
          ↓
   Stack Overflow
          ↓
          AI
          ↓
  git reset --hard
          ↓
    Wait... WHAT?
```

---

## The Idea

While you type a Git command, Gogit understands the current context
and suggests available options.

For example:

```bash
git branch --
```

Gogit may show:

```text
--show-current
    Show the name of the current branch.

--merged
    List branches already merged into the specified commit.

--no-merged
    List branches that have not yet been merged.

--delete
    Delete a branch.
```

So instead of only telling you:

> **what you can type**

Gogit also tells you:

> **what the hell it actually does**

For dangerous commands, Gogit should eventually be able to tell you
that you're about to do something... interesting:

```text
--force
    Force push to the remote repository.

    ⚠ This may overwrite remote history.

--force-with-lease
    Force push only when the remote branch has not unexpectedly changed.

    ✓ Usually safer than --force.
```

Because sometimes the most useful feature of a CLI assistant
is not helping you type:

```bash
git reset --hard
```

faster.

It's stopping you for half a second before you do it.

---

## Why "Gogit"?

Because it's written in **Go**.

And it's for **Git**.

Go + Git.

**Gogit.**

Yes.

I spent considerably more time Googling Git commands than naming this project.

> [!Warning]
> If you use agent to operate git, bro, this is not what you need.
> just use you agent and abandon your brain. Be a nerd only use AI to code and commit.
> (I'm not mean AI is harmful, I just think we should not let AI do everything, so that we won't forget something basic)