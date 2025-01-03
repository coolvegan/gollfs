# Ollama Failsafe
The need to have more than one ollama addresses for my stuff 
motivated e to write Ollama failsafe

## Configuration
watchdog=no
\# Interval in seconds
interval=60
\# How long the http.Client should wait until decision that server is not running
\# Timeout in milliseconds
timeout=500
\# Ollama Server, priority, less means higher priority, should be the strongest pc
server=http://localhost:11434,5
server=http://10.0.0.8:8000,5
