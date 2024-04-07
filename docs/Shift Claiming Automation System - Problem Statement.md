### Problem Statement

#### Overview

The goal of this project is to automate the process of claiming shifts through an online portal for an on-campus job. This automation is intended to monitor available shifts in real-time and claim them automatically, enhancing efficiency and ensuring timely shift allocation.

#### Technical Description

- **Endpoint Interaction:** The online portal exposes an endpoint that, when accessed via GET requests, returns a list of available shifts. A separate endpoint allows for claiming these shifts through POST requests, using the shift ID obtained from the GET request.
- **Polling Mechanism:** To monitor shift availability, continuous GET requests are made to the shift listing endpoint at a frequency of every five seconds. This polling ensures real-time detection of new shifts.
- **Shift Claiming:** Upon detecting an available shift, a POST request is made to the claiming endpoint with the relevant shift ID, thus automating the process of shift allocation.
- **Error Handling and Timeout:** The system incorporates a mechanism to handle timeouts, which are imposed after a certain number of unsuccessful requests or due to server-defined limitations. In the event of a timeout, the system is designed to halt all requests for a 30-minute cooldown period, after which polling can resume.

#### Constraints and Considerations

- **Timeout Handling:** Critical to the system is an efficient method for managing the 30-minute timeout period, necessitating a pause in the polling process to comply with the server's request rate limitations.
- **Current Implementation:** The existing solution operates on a virtual machine (VM), utilizing a Python script that adheres to the described workflow, including a static 30-minute sleep period for cooldown management.
- **Objective:** The project seeks to transition to a more cloud-native approach, optimizing for scalability, reliability, and compliance with the online portal's operational constraints. This includes exploring more dynamic and resource-efficient methods for polling and timeout management, moving away from the static, VM-based setup.

#### Additional Requirements

- **Manual Control:** The system should allow for manual start and stop operations, enabling the user to control the polling and claiming process as needed.
- **Monitoring and Observability:** Implement monitoring and observability features to track the system's operation, including successful and failed requests, timeouts, and overall system health.
- **Usage Metrics:** The solution should provide insights into usage patterns, including the number of shifts claimed, frequency of timeouts, and resource consumption, to optimize performance and manage resources efficiently.

#### Goals

1. Develop a cloud-native solution that automates the process of monitoring and claiming shifts through the portal's endpoints.
2. Implement a more dynamic and efficient cooldown management system to handle timeouts without resorting to prolonged, static sleep intervals.
3. Ensure the solution is scalable, reliable, and can operate within the constraints imposed by the online portal's API, including request rate limitations and timeout penalties.

This problem statement outlines the requirements and constraints for automating shift claims through an online portal, providing a foundation for exploring potential cloud-native solutions that address the identified challenges.
