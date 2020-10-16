# job-observer
![Go test](https://github.com/akaimo/job-observer/workflows/Go%20test/badge.svg)

job-observer deletes a Job that has been stopped for an arbitrary amount of time.

# Install

```
$ kubectl apply -f https://raw.githubusercontent.com/akaimo/job-observer/master/bundle.yaml
```

# Configuration
## Cleaner

Specify the job to be deleted in the Cleaner.

```
apiVersion: job-observer.akaimo.com/v1alpha1
kind: Cleaner
metadata:
  name: my-cleaner
spec:
  # It will be deleted after this time
  # Specify time as "s", "m", "h"
  ttlAfterFinished: 1h
  
  # Status of the job to be deleted when the job is stopped
  # "Complete" or "Failed" or ""
  # Default is ""
  # The "" applies to all Jobs
  cleaningJobStatus: Complete
  
  # The job that matches the label is deleted.
  selector:
    matchLabels:
      app: my
      version: mikan
```
