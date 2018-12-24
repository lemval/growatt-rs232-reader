HOST=localhost
PORT=5701

curl -s http://$HOST:$PORT/status | json_pp --json_opt=canonical,pretty
curl -s http://$HOST:$PORT/info | json_pp --json_opt=canonical,pretty
