# go-queue

The right way to implement background workers.


Application-logic is done at the `usecase` layer. However, there are times where we want to run the process in the background. Registering such process should be trivial, and you shouldn't need to know about the details of the background worker.
