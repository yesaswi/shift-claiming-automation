# Shift Claiming Automation System

## Overview

The Shift Claiming Automation System is a serverless solution built on Google Cloud that automates the process of monitoring and claiming shifts through an online portal's API. The system utilizes a Cloud Run service to handle the core logic of fetching shifts, claiming them, and managing the start/stop functionality. It leverages Cloud Tasks for scheduling tasks and Cloud Pub/Sub for receiving start/stop commands. The system state is stored in Firestore, and the architecture is designed to handle cooldown periods efficiently.

## Architecture

The system architecture consists of the following components:

1. **Shift Claiming Service**: A Go program deployed as a Cloud Run service handles the core logic of polling the shift listing endpoint, checking for available shifts, and claiming them if found. It also includes endpoints for the start/stop functionality. The service is triggered by Cloud Tasks based on a scheduled interval and by Cloud Pub/Sub for the start/stop commands.

2. **Cloud Tasks**: Google Cloud Tasks is used to schedule and enqueue tasks for the Shift Claiming Service. It triggers the service at a specified interval and handles task retries and error handling.

3. **Cloud Pub/Sub**: Google Cloud Pub/Sub is used to enable the manual start/stop functionality. It allows the system to receive start/stop commands and triggers the corresponding endpoints in the Shift Claiming Service.

4. **Firestore**: Google Cloud Firestore, a NoSQL document database, is used to store the system state, including the start/stop flag and other relevant data. It provides real-time synchronization and scalability.

5. **Cloud Logging and Monitoring**: Google Cloud's logging and monitoring services are utilized to track the system's operation, including successful and failed requests, timeouts, start/stop commands, and overall system health.

## Workflow

1. **Manual Start/Stop**:

   - Users can send "start" or "stop" commands via a Cloud Pub/Sub topic.
   - The Shift Claiming Service receives the command through a Pub/Sub subscription.
   - It sets the start/stop flag in Firestore accordingly.
   - The service manages the scheduling of tasks in Cloud Tasks based on the start/stop state.

2. **Task Scheduling**:

   - When the system is in the "start" state, Cloud Tasks schedules tasks for the Shift Claiming Service at a specified interval (e.g., every 5 seconds).
   - Each task represents a request to fetch and claim shifts.
   - Cloud Tasks enqueues the tasks and triggers the Shift Claiming Service accordingly.

3. **Shift Claiming Service Execution**:

   - The Shift Claiming Service is triggered by Cloud Tasks for each scheduled task.
   - It checks the start/stop flag in Firestore during each invocation.
   - If the start/stop flag is set to "start", the service proceeds with its normal operation.
   - It makes a GET request to the shift listing endpoint of the online portal's API to retrieve the list of available shifts.
   - If an available shift is found, the service immediately makes a POST request to the claiming endpoint to claim the shift.
   - The service stores the result of the claiming process (success or failure) in Firestore.

4. **Cooldown Mechanism**:

   - If a timeout or rate limiting response is encountered during the shift listing or claiming process, the Shift Claiming Service initiates the cooldown mechanism.
   - It calculates the next invocation time by adding the cooldown duration to the current timestamp.
   - The service schedules the next task in Cloud Tasks to run at the calculated next invocation time.
   - This ensures that the service respects the cooldown period and avoids making unnecessary requests during that time.

5. **Resuming Normal Operation**:

   - Once the cooldown period has expired and the next invocation time is reached, the Shift Claiming Service resumes its normal operation, provided the start/stop flag is set to "start".
   - The service continues to schedule tasks in Cloud Tasks at the specified interval.

6. **Monitoring and Logging**:
   - Cloud Logging captures logs from the Shift Claiming Service, including request details, claimed shifts, timeouts, start/stop commands, and any errors encountered.
   - Cloud Monitoring tracks key metrics, such as request latency, success rates, error counts, and the current status of the automation process (started or stopped).
   - Alerts and notifications can be configured based on predefined thresholds or anomalies detected in the monitored metrics.

## Firestore Data Model

The Firestore data model for storing the system state is structured as follows:

- **Configuration Collection**:

  - Document ID: "config"
  - Fields:
    - "startStopFlag": boolean (true for "start", false for "stop")

- **ClaimingResults Collection**:
  - Document ID: auto-generated
  - Fields:
    - "shiftId": string (the ID of the claimed shift)
    - "claimingStatus": string ("success" or "failure")
    - "timestamp": timestamp (the timestamp of the claiming attempt)

This Firestore data model allows for efficient storage and retrieval of the system state, including the start/stop flag and claiming results. The Configuration collection holds the global configuration settings, while the ClaimingResults collection stores the history of claiming attempts.

## Scalability and Reliability

The Shift Claiming Automation System is designed to be scalable and reliable:

- **Cloud Run**: Cloud Run automatically scales the Shift Claiming Service based on the incoming requests, ensuring that the service can handle varying levels of traffic.

- **Cloud Tasks**: Cloud Tasks provides a reliable and scalable task scheduling mechanism. It ensures that tasks are executed reliably, with built-in retries and error handling.

- **Firestore**: Firestore offers automatic scaling, real-time synchronization, and strong consistency. It can handle a high volume of reads and writes, making it suitable for storing the system state.

- **Pub/Sub**: Cloud Pub/Sub is a highly scalable and reliable messaging service. It ensures that start/stop commands are delivered reliably to the Shift Claiming Service.

## Error Handling and Resilience

The system incorporates error handling and resilience mechanisms to handle failures gracefully:

- **Retries**: Cloud Tasks automatically retries failed tasks based on a configured retry policy. This helps to handle transient failures and ensures that tasks are eventually executed successfully.

- **Error Logging**: Any errors encountered during the execution of the Shift Claiming Service are logged using Cloud Logging. This enables quick identification and troubleshooting of issues.

- **Monitoring and Alerts**: Cloud Monitoring allows for setting up alerts based on predefined thresholds or anomalies. Alerts can be configured to notify the relevant team members in case of critical errors or performance degradation.

## Conclusion

The Shift Claiming Automation System provides a scalable, reliable, and efficient solution for automating the process of monitoring and claiming shifts through an online portal's API. By leveraging Cloud Run, Cloud Tasks, Firestore, and Pub/Sub, the system can handle the core functionality of fetching and claiming shifts, manage start/stop commands, and handle cooldown periods effectively.

The architecture is designed to be resilient, with error handling and monitoring mechanisms in place to ensure smooth operation. The use of Firestore for state management and Pub/Sub for start/stop commands enables real-time synchronization and reliable communication between components.

This design document provides a comprehensive overview of the Shift Claiming Automation System, detailing its architecture, workflow, data model, scalability, reliability, and error handling aspects. It serves as a blueprint for implementing and deploying the system on the Google Cloud Platform.
