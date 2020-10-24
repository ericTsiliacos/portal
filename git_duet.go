package main

func gitDuet() []string {
	author := execute("git config --get duet.env.git-author-initials")
	coauthor := execute("git config --get duet.env.git-committer-initials")

	return []string{author, coauthor}
}
