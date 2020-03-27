package main

import (
	"fmt"
	"time"

	"github.com/go-git/go-git"
	"github.com/go-git/go-git/plumbing/object"
)

func testGit() {

	// r, err := git.PlainClone("/tmp/foo", false, &git.CloneOptions{
	// 	URL:      "https://github.com/romanprog/hiver",
	// 	Progress: os.Stdout,
	// })
	path := "/tmp/foo"
	r, err := git.PlainOpen(path)
	checkErr(err)
	// Gets the HEAD history from HEAD, just like this command:
	log.Info("git log")
	// ... retrieves the branch pointed by HEAD
	ref, err := r.Head()
	checkErr(err)

	w, err := r.Worktree()
	checkErr(err)
	log.Info("Print")
	for {
		err = w.Pull(&git.PullOptions{RemoteName: "origin"})
		if err != nil {
			log.Info(err.Error())
			time.Sleep(time.Second * 30)
			continue
		}

		// Print the latest commit that was just pulled
		ref, err := r.Head()
		checkErr(err)
		commit, err := r.CommitObject(ref.Hash())
		checkErr(err)
		fmt.Println(commit)

		cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
		checkErr(err)

		// ... just iterates over the commits, printing it
		err = cIter.ForEach(func(c *object.Commit) error {
			fmt.Println(c)
			return nil
		})

		time.Sleep(time.Second * 30)
	}
	// ... retrieves the commit history
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})

	checkErr(err)
	// ... just iterates over the commits, printing it
	err = cIter.ForEach(func(c *object.Commit) error {
		fmt.Println(c)
		return nil
	})
	checkErr(err)
}
