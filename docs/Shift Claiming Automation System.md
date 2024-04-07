# Shift Claiming Automation System

## Overview

The Shift Claiming Automation System is a serverless solution built on Google Cloud that automates the process of monitoring and claiming shifts through an online portal's API. The system dynamically schedules the execution of a Cloud Function to poll the shift listing endpoint, check for available shifts, and automatically claim them by making POST requests to the claiming endpoint. It incorporates an optimized cooldown mechanism to handle timeouts and rate limiting while minimizing unnecessary resource usage. Additionally, it provides a manual start/stop functionality to control the automation process on-demand.

## Architecture

The system architecture consists of the following components:

1. **Shift Claiming Function**: A Cloud Function, written in Golang, handles the core logic of polling the shift listing endpoint, checking for available shifts, and claiming them if found. It is triggered by Cloud Scheduler based on a dynamic invocation schedule.

2. **Start/Stop Function**: A separate Cloud Function, triggered by Cloud Pub/Sub or an API endpoint, handles the manual start/stop functionality. It sets the start/stop flag in Cloud Memorystore or Cloud Firestore and sends messages to the Cloud Scheduler to control the scheduling of the Shift Claiming Function.

3. **Cloud Scheduler**: Google Cloud Scheduler is used to trigger the Shift Claiming Function based on a dynamic invocation schedule. The schedule is adjusted dynamically to optimize resource usage during cooldown periods and is controlled by messages from the Start/Stop Function.

4. **Cloud Memorystore (Redis) or Cloud Firestore**: Cloud Memorystore (if using Redis) or Cloud Firestore (if using a document database) is used to store the cooldown flag, next invocation time, start/stop flag, and other relevant data. It provides fast and reliable storage for managing the system state.

5. **Cloud Pub/Sub**: Google Cloud Pub/Sub is used to enable the manual start/stop functionality. It allows the system to receive start/stop commands and triggers the Start/Stop Function.

6. **Cloud Logging and Monitoring**: Google Cloud's logging and monitoring services are utilized to track the system's operation, including successful and failed requests, timeouts, start/stop commands, and overall system health.

## Workflow

1. **Manual Start/Stop**:

   - Users can send "start" or "stop" commands via an API endpoint or Cloud Pub/Sub topic.
   - The Start/Stop Function is triggered by the received command.
   - It sets the start/stop flag in Cloud Memorystore or Cloud Firestore accordingly.
   - The Start/Stop Function sends a message to the Cloud Scheduler to resume or pause the scheduling of the Shift Claiming Function.

2. **Dynamic Scheduling**:

   - The Cloud Scheduler triggers the Shift Claiming Function based on the dynamic invocation schedule when the automation process is in the "start" state.
   - When a cooldown is initiated, the Shift Claiming Function calculates the next invocation time by adding the cooldown duration to the current timestamp.
   - It updates the Cloud Scheduler's configuration to schedule the next invocation at the specified next invocation time.

3. **Shift Claiming Function Execution**:

   - The Shift Claiming Function checks the start/stop flag in Cloud Memorystore or Cloud Firestore during each invocation.
   - If the start/stop flag is set to "start", the function proceeds with its normal operation.
   - It makes a GET request to the shift listing endpoint of the online portal's API to retrieve the list of available shifts.
   - If an available shift is found, the function immediately makes a POST request to the claiming endpoint to claim the shift.
   - The function stores the result of the claiming process (success or failure) in Cloud Memorystore or Cloud Firestore.

4. **Cooldown Mechanism**:

   - If a timeout or rate limiting response is encountered during the shift listing or claiming process, the Shift Claiming Function initiates the cooldown mechanism.
   - It sets a cooldown flag and calculates the next invocation time by adding the cooldown duration to the current timestamp.
   - The cooldown flag and next invocation time are stored in Cloud Memorystore or Cloud Firestore.
   - The function updates the Cloud Scheduler's configuration to schedule the next invocation at the specified next invocation time.

5. **Early Termination During Cooldown**:

   - When the Cloud Scheduler triggers the Shift Claiming Function during the cooldown period, the function checks the cooldown flag and next invocation time stored in Cloud Memorystore or Cloud Firestore.
   - If the cooldown flag is present and the current time is earlier than the next invocation time, the function terminates early without making any API requests.

6. **Resuming Normal Operation**:

   - Once the cooldown period has expired and the next invocation time is reached, the Shift Claiming Function resumes its normal operation, provided the start/stop flag is set to "start".
   - The cooldown flag and next invocation time are removed from Cloud Memorystore or Cloud Firestore.
   - The function updates the Cloud Scheduler's configuration to revert back to the original fixed interval (e.g., every 5 seconds) for subsequent invocations.

7. **Monitoring and Logging**:
   - Cloud Logging captures logs from the Shift Claiming Function and Start/Stop Function, including request details, claimed shifts, timeouts, start/stop commands, and any errors encountered.
   - Cloud Monitoring tracks key metrics, such as request latency, success rates, error counts, and the current status of the automation process (started or stopped).
   - Alerts and notifications can be configured based on predefined thresholds or anomalies detected in the monitored metrics.

## Cooldown Mechanism Details

The optimized cooldown mechanism handles timeouts and rate limiting while minimizing unnecessary resource usage:

1. **Cooldown Initiation**:

   - When a timeout or rate limiting response is encountered, the Shift Claiming Function initiates the cooldown mechanism.
   - It sets a cooldown flag and calculates the next invocation time by adding the cooldown duration to the current timestamp.
   - The cooldown flag and next invocation time are stored in Cloud Memorystore or Cloud Firestore.

2. **Dynamic Scheduling**:

   - The Shift Claiming Function updates the Cloud Scheduler's configuration to schedule the next invocation at the specified next invocation time.
   - This ensures that the function is not triggered unnecessarily during the cooldown period, reducing resource usage and costs.

3. **Early Termination During Cooldown**:

   - When the Cloud Scheduler triggers the Shift Claiming Function during the cooldown period, the function checks the cooldown flag and next invocation time.
   - If the cooldown flag is present and the current time is earlier than the next invocation time, the function terminates early without making any API requests.

4. **Resuming Normal Operation**:
   - Once the cooldown period has expired and the next invocation time is reached, the Shift Claiming Function resumes its normal operation, provided the start/stop flag is set to "start".
   - The cooldown flag and next invocation time are removed from Cloud Memorystore or Cloud Firestore.
   - The function updates the Cloud Scheduler's configuration to revert back to the original fixed interval for subsequent invocations.

## Manual Start/Stop Functionality

The manual start/stop functionality allows users to control the automation process on-demand:

1. **Start Command**:

   - When a "start" command is received via the API endpoint or Cloud Pub/Sub topic, the Start/Stop Function is triggered.
   - It sets the start/stop flag to "start" in Cloud Memorystore or Cloud Firestore.
   - The function sends a message to the Cloud Scheduler to resume the scheduling of the Shift Claiming Function at the specified interval.

2. **Stop Command**:
   - When a "stop" command is received via the API endpoint or Cloud Pub/Sub topic, the Start/Stop Function is triggered.
   - It sets the start/stop flag to "stop" in Cloud Memorystore or Cloud Firestore.
   - The function sends a message to the Cloud Scheduler to pause the scheduling of the Shift Claiming Function.

This comprehensive documentation covers the end-to-end architecture and workflow of the Shift Claiming Automation System. It incorporates the optimized cooldown mechanism, dynamic scheduling, manual start/stop functionality, and monitoring capabilities to create a robust and efficient solution for automating the shift claiming process.

Please note that this documentation provides a high-level overview of the system. The actual implementation may involve additional details, error handling, and best practices specific to the chosen programming language and Google Cloud services.
