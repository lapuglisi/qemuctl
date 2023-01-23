package qemuctl_helpers

type QemuctlCompletion struct {
	shellName string
}

func (q *QemuctlCompletion) GetCompletion(shell string) string {
	var completionString string = ""

	q.shellName = shell

	switch q.shellName {
	case "zsh":
		{
			completionString = q.zshCompletion()
		}
	case "bash":
		{
			completionString = q.bashCompletion()
		}
	default:
		{

		}
	}

	return completionString
}

func (q *QemuctlCompletion) zshCompletion() string {
	return `# asdasd
function _qemuctl() {
	local -a qemuctl_actions=(list start stop destroy create status edit);
	local -a qemuctl_machines=( $(qemuctl list --no-headings --names-only) );

	case $CURRENT in
	2)
		compadd -a qemuctl_actions;
	;;

	3)
		compadd -a qemuctl_machines;
	;;

	*)
		_files "*";
	;;
	esac
}

compctl -K _qemuctl qemuctl;

#_qemuctl "$@"
`
}

func (q *QemuctlCompletion) bashCompletion() string {
	return "not implemented"
}
