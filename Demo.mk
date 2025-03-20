localPort?=8081
demo-rev-port-forward:
	while true; do \
		ssh -p 2222 \
		-o UserKnownHostsFile=/dev/null \
		-o StrictHostKeyChecking=no \
		-NR 8083:localhost:8081 \
		localhost; \
		sleep 5; \
		done

demo-port-forward:
	while true; do \
		ssh -p 2222 \
		-o UserKnownHostsFile=/dev/null \
		-o StrictHostKeyChecking=no \
		-NL 8082:localhost:8083 \
		localhost; \
		sleep 5; \
		done

demo-local-service:
	wfsd -p :$(localPort)

demo-call-port-forward:
	curl http://localhost:8082
