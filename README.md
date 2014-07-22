scale-event-trigger
===================

Runs a command when EC2 instances that match the given tags change.

Specify tags to limit instances that are checked, multiple tags can be used. You must specify at least one tag.

`./scale-event-trigger Service:web Environment:live`

Set the command to run with the --command flag:

`./scale-event-trigger --command="command to run here"`

scale-event-trigger will check the EC2 instances once every 60 seconds, you can change this as follows:

`./scale-event-trigger --frequency=120`

Authentication
--------------

AWS access key and secret key can be added as environment variables, using either `AWS_ACCESS_KEY_ID` or `AWS_ACCESS_KEY` and `AWS_SECRET_ACCESS_KEY` or `AWS_SECRET_KEY`.  If these are not available then IAM credentials for the instance will be checked.

Building
--------

Build a binary by running:

`go build scale-event-trigger.go`
