# Crow

Crow is a tool similar to Tempo Vulture and Loki Canary that is used to smoke test Grafana Agent. Crow works by generating metrics, then validating them against Prometheus. Crow uses two endpoints; the traditional `/metrics` and then `/validate` that generates the results of Crow checking for successful samples. 
 



