#!/bin/env bash


command -v tmux > /dev/null
if [[ $? -eq 1 ]]; then
	echo "TMUX must be installed to run this script"
	exit 1
fi


# Start TMUX session if there is none
tmux list-sessions > /dev/null
NEW_SESSION=0
if [[ $? -eq 1 ]]; then
	NEW_SESSION=1
	tmux new-session -d -s lanchat
fi

# Get current window
active_window=$(tmux list-windows -F "#{window_active} #{window_id}" | awk '$1 == 1 { print $2 }')
multi_pane=$(tmux list-panes | wc -l)
active_pane=$(tmux list-panes -F "#{pane_active} #{pane_index}" | awk '$1 == 1 { print $2 }')


cleanup() {
	if [[ $multi_pane -gt 1 ]] ; then
		tmux join-pane -t "${active_window}.right" 2>/dev/null
		sleep 0.1
	fi
	if [[ $NEW_SESSION -eq 1 ]]; then
		tmux kill-session -t lanchat 2>/dev/null
	else
		tmux kill-window -t fake_chat 2>/dev/null
	fi
        exit 0
}
trap cleanup SIGHUP SIGINT SIGTERM

tmux new-window -d -n fake_chat "bin/lanchat -u anna"
sleep 0.1
tmux split-window -h -t fake_chat "bin/lanchat -u bob"
sleep 0.1
tmux split-window -v -t fake_chat.right "bin/lanchat -u conan"
sleep 0.1

tmux send-keys -t fake_chat.left "hello, my friend" Enter
sleep 0.1
tmux send-keys -t fake_chat.top-right "wie geht's?" Enter
sleep 0.1
tmux send-keys -t fake_chat.bottom-right "Hey hey!" Enter
sleep 0.1
tmux send-keys -t fake_chat.bottom-right ":id zonan" Enter
sleep 0.1
tmux send-keys -t fake_chat.bottom-right "I'm Zonan now" Enter
sleep 0.1
tmux send-keys -t fake_chat.top-right ":fake cmd" Enter
sleep 0.1

if [[ $multi_pane -gt 1 ]] ; then
	tmux move-pane -t fake_chat.bottom-left
fi
tmux select-window -t fake_chat

read -n 1 -s -r -p "Press any key to close"
echo ""
cleanup
