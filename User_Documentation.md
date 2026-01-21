# User Documentation: EMSIB Platform

**Enterprise Microservices Infrastructure for Buildings**

---

## Document Information

| Field | Value |
|-------|-------|
| **Document Type** | User Documentation |
| **Version** | 1.0 |
| **Date** | 2024 |
| **Platform** | EMSIB - Enterprise Microservices Infrastructure for Buildings |

---

## Table of Contents

1. [Introduction](#1-introduction)
2. [Target Users](#2-target-users)
3. [System Overview (from User Perspective)](#3-system-overview-from-user-perspective)
4. [Functional Capabilities](#4-functional-capabilities)
5. [How to Use the System / Component](#5-how-to-use-the-system--component)
6. [Error Handling & System Behavior](#6-error-handling--system-behavior)
7. [Best Practices & Recommendations](#7-best-practices--recommendations)
8. [Limitations](#8-limitations)
9. [Conclusion](#9-conclusion)

---

## 1. Introduction

### What is EMSIB?

**EMSIB (Enterprise Microservices Infrastructure for Buildings)** is a comprehensive smart building management platform designed to help organizations monitor, control, and optimize their building energy systems. The platform enables real-time monitoring of IoT devices, automated energy optimization, predictive analytics, and intelligent decision-making for building operations.

### What Problem Does It Solve?

Modern buildings face numerous challenges in energy management:

- **High Energy Costs**: Buildings consume significant amounts of energy, leading to substantial operational costs
- **Peak Demand Charges**: Utility companies charge premium rates during peak consumption periods
- **Device Management Complexity**: Managing hundreds of IoT devices (HVAC systems, sensors, lighting, meters) manually is impractical
- **Lack of Visibility**: Building managers often lack real-time insights into energy consumption patterns and anomalies
- **Reactive Maintenance**: Issues are discovered only after they cause problems, leading to downtime and inefficiency

### What Value Does It Provide?

EMSIB provides tangible value to users:

- **Cost Reduction**: Automated optimization reduces energy consumption and peak demand charges
- **Operational Efficiency**: Real-time monitoring and automated control reduce manual intervention
- **Predictive Insights**: Forecasting capabilities help anticipate energy demand and plan accordingly
- **Anomaly Detection**: Early detection of unusual consumption patterns prevents equipment failures
- **Data-Driven Decisions**: Comprehensive analytics and reports support informed decision-making
- **Compliance Support**: Detailed audit logs and reporting help meet regulatory requirements

### How It Fits into the Whole System

EMSIB operates as an integrated platform consisting of four core microservices that work together:

1. **Security Service**: Provides authentication and authorization for all users and services
2. **IoT Control Service**: Manages all IoT devices, collects telemetry data, and executes control commands
3. **Forecast Service**: Generates energy demand predictions and optimization scenarios
4. **Analytics Service**: Processes data to provide insights, reports, KPIs, and anomaly detection

These services communicate seamlessly to provide a unified experience. For example, when you request a building dashboard, the Analytics Service automatically gathers device data from the IoT Service and forecast data from the Forecast Service to present a comprehensive view.

---

## 2. Target Users

EMSIB serves multiple user types, each with different needs and access levels:

### 2.1 End Users (Human Users)

#### Building Managers
- **Who they are**: Facility managers, building operators, and property managers responsible for day-to-day building operations
- **What they do**: Monitor building performance, manage devices, review reports, and make operational decisions
- **Key needs**: Real-time device status, energy consumption reports, anomaly alerts, device control capabilities

#### Energy Analysts
- **Who they are**: Energy consultants, sustainability officers, and analysts focused on energy efficiency
- **What they do**: Analyze energy consumption patterns, generate reports, identify optimization opportunities
- **Key needs**: Historical data access, detailed analytics, forecasting information, comparative reports

#### System Administrators
- **Who they are**: IT administrators and system operators managing the EMSIB platform
- **What they do**: Manage user accounts, configure roles and permissions, monitor system health, maintain audit logs
- **Key needs**: User management, role configuration, system monitoring, audit log access

#### Facility Technicians
- **Who they are**: Maintenance staff and technicians who work with physical devices
- **What they do**: Respond to alerts, perform maintenance, verify device status
- **Key needs**: Device status information, anomaly notifications, command execution capabilities

### 2.2 System Administrators

#### Platform Administrators
- **Who they are**: Technical staff responsible for deploying, configuring, and maintaining the EMSIB platform infrastructure
- **What they do**: Configure services, manage databases, monitor system health, handle deployments
- **Key needs**: Service configuration, health monitoring, database management, deployment tools

### 2.3 Other System Components / Services (Backend Users)

#### External Systems
- **Who they are**: Third-party applications, building management systems (BMS), or enterprise resource planning (ERP) systems
- **What they do**: Integrate with EMSIB to exchange data, receive notifications, or trigger actions
- **Key needs**: API access, webhook support, data synchronization

#### IoT Devices
- **Who they are**: Physical devices (HVAC systems, sensors, smart meters, lighting controls)
- **What they do**: Send telemetry data, receive commands, report status
- **Key needs**: MQTT communication, command execution, status reporting

---

## 3. System Overview (from User Perspective)

### What the System Allows Users to Do

EMSIB enables users to:

1. **Monitor Building Systems in Real-Time**
   - View live status of all IoT devices (HVAC, sensors, lighting, meters)
   - Monitor energy consumption, temperature, humidity, and other metrics
   - Track device health and connectivity status

2. **Control IoT Devices**
   - Send commands to devices (e.g., adjust temperature, dim lighting, change modes)
   - Execute optimization scenarios automatically
   - Schedule device actions based on forecasts

3. **Generate and View Forecasts**
   - Request energy demand predictions for buildings or devices
   - Identify peak load periods
   - Receive optimization recommendations

4. **Access Analytics and Reports**
   - View comprehensive dashboards showing building performance
   - Generate detailed reports on energy consumption, costs, and efficiency
   - Review Key Performance Indicators (KPIs)
   - Monitor and acknowledge anomalies

5. **Manage Users and Permissions**
   - Create and manage user accounts
   - Assign roles and permissions
   - Review audit logs of system activities

### Main Responsibilities of the Component

The EMSIB platform is responsible for:

- **Device Management**: Registering, tracking, and managing all IoT devices in buildings
- **Data Collection**: Continuously collecting telemetry data from devices via HTTP and MQTT
- **Data Processing**: Analyzing telemetry data to detect anomalies, calculate KPIs, and generate insights
- **Forecasting**: Predicting future energy demand using historical data, weather information, and machine learning
- **Optimization**: Generating and executing scenarios to reduce energy consumption and costs
- **Reporting**: Creating comprehensive reports and dashboards for decision-making
- **Security**: Authenticating users, authorizing actions, and maintaining audit trails

### How Users Interact with It

Users interact with EMSIB through multiple interfaces:

#### 1. REST API (Primary Interface)
- **HTTP/HTTPS requests**: Users send HTTP requests to various endpoints
- **JSON format**: Data is exchanged in JSON format
- **Authentication**: All requests require JWT (JSON Web Token) authentication
- **Example**: A building manager sends a GET request to retrieve device status

#### 2. MQTT Protocol (For IoT Devices)
- **Real-time messaging**: Devices communicate via MQTT broker
- **Bidirectional**: Devices send telemetry and receive commands
- **Automatic**: Handled automatically by the IoT Control Service

#### 3. Web Dashboard (Future)
- **Graphical interface**: Visual dashboards for non-technical users
- **Interactive charts**: Real-time graphs and visualizations
- **Point-and-click**: User-friendly interface for common operations

#### 4. Command Line Interface (For Administrators)
- **Scripting support**: Automation and integration capabilities
- **Bulk operations**: Efficient handling of multiple operations

---

## 4. Functional Capabilities

### 4.1 Authentication and User Management

#### Login and Token Management
- **Login**: Users authenticate with username and password to receive access tokens
- **Token Refresh**: Refresh expired tokens without re-entering credentials
- **User Information**: Retrieve current user profile and permissions
- **Logout**: Invalidate tokens and end sessions

#### User Account Management (Admin Only)
- **Create Users**: Add new user accounts with roles and permissions
- **View Users**: List all users in the system
- **Update Users**: Modify user information, roles, and status
- **Delete Users**: Remove user accounts from the system

#### Role and Permission Management (Admin Only)
- **Create Roles**: Define custom roles with specific permissions
- **Assign Permissions**: Grant access to resources (buildings, devices, reports) and actions (read, write, delete)
- **Manage Roles**: Update or delete role definitions

### 4.2 Device Management

#### Device Registration
- **Register Devices**: Add new IoT devices to the system
- **Device Information**: Specify device type, model, location, and capabilities
- **Location Tracking**: Associate devices with buildings, floors, and rooms

#### Device Monitoring
- **List Devices**: View all devices, filtered by building, type, or status
- **Device Details**: Retrieve comprehensive information about specific devices
- **Device Status**: Monitor online/offline status and last seen timestamps
- **Live State**: View real-time device states and latest telemetry

### 4.3 Telemetry Data Management

#### Data Ingestion
- **Single Telemetry**: Send individual telemetry readings from devices
- **Bulk Telemetry**: Efficiently send multiple telemetry readings at once
- **Automatic Collection**: Devices can send data via MQTT automatically

#### Data Retrieval
- **Historical Data**: Query telemetry history for specific devices
- **Time Range Queries**: Retrieve data for specific time periods
- **Filtering**: Filter by device, metric, or time range

### 4.4 Device Control

#### Command Execution
- **Send Commands**: Issue commands to devices (e.g., SET_TEMPERATURE, SET_MODE, TURN_OFF)
- **Command Status**: Track command execution status (PENDING, SENT, APPLIED, FAILED)
- **Command History**: View past commands and their outcomes
- **Real-time Execution**: Commands are sent immediately via MQTT

### 4.5 Forecasting

#### Energy Demand Forecasting
- **Generate Forecasts**: Create energy demand predictions for buildings or devices
- **Forecast Types**: Demand, consumption, or load profile forecasts
- **Time Horizons**: Forecast from 1 hour to 7 days ahead
- **Weather Integration**: Include weather data for improved accuracy
- **Tariff Integration**: Consider energy pricing for cost optimization

#### Peak Load Prediction
- **Identify Peaks**: Predict when peak energy consumption will occur
- **Threshold Alerts**: Receive warnings when peaks exceed thresholds
- **Recommendations**: Get suggestions for peak shaving strategies

#### Device-Level Predictions
- **Device Forecasts**: Get consumption predictions for individual devices
- **Trend Analysis**: Understand consumption trends (increasing, decreasing, stable)
- **Optimization Insights**: Receive device-specific optimization recommendations

### 4.6 Optimization Scenarios

#### Scenario Generation
- **Create Scenarios**: Generate optimization scenarios based on forecasts
- **Optimization Types**: 
  - Cost reduction
  - Peak shaving
  - Load balancing
  - Efficiency improvement
  - Comfort optimization
  - Demand response
- **Expected Savings**: View predicted energy, cost, and CO2 savings
- **Constraints**: Set limits (e.g., minimum/maximum temperature, preserve comfort)

#### Scenario Execution
- **Apply Scenarios**: Send approved scenarios to IoT Control Service for execution
- **Progress Tracking**: Monitor execution progress in real-time
- **Action Status**: Track status of individual actions within scenarios
- **Dry Run**: Test scenarios without actually executing commands

### 4.7 Analytics and Reporting

#### Dashboards
- **Overview Dashboard**: System-wide view of all buildings, devices, and KPIs
- **Building Dashboard**: Detailed view for specific buildings
- **Real-time Updates**: Current status of devices, consumption, and anomalies
- **Forecast Integration**: View forecast data alongside current metrics

#### Reports
- **Report Generation**: Create detailed reports on:
  - Energy consumption
  - Cost analysis
  - Efficiency metrics
  - Sustainability (carbon footprint)
  - Period comparisons
- **Report Types**: Pre-defined templates for common analysis needs
- **Custom Options**: Include breakdowns, recommendations, and comparisons
- **Report History**: Access previously generated reports

#### Key Performance Indicators (KPIs)
- **System KPIs**: Overall system performance metrics
- **Building KPIs**: Building-specific metrics including:
  - Energy intensity
  - Peak demand ratio
  - Equipment efficiency
  - Renewable energy share
  - Carbon intensity
  - Cost per square meter
- **Period-based**: Daily, weekly, or monthly KPI calculations
- **Manual Calculation**: Trigger KPI recalculation on demand

#### Anomaly Detection
- **Automatic Detection**: System automatically detects unusual consumption patterns
- **Anomaly Types**: Consumption spikes, unusual patterns, threshold violations
- **Severity Levels**: LOW, MEDIUM, HIGH, CRITICAL
- **Anomaly Management**: 
  - View detected anomalies
  - Acknowledge anomalies
  - Mark as resolved or false positive
- **Filtering**: Filter by device, building, severity, or status

#### Time-Series Analysis
- **Query Time-Series Data**: Retrieve aggregated time-series data
- **Aggregation Types**: Average, sum, minimum, maximum, count
- **Flexible Intervals**: Hourly, daily, or custom intervals
- **Multiple Metrics**: Query different metrics (temperature, consumption, etc.)

### 4.8 Audit and Compliance

#### Audit Logging
- **Automatic Logging**: All significant actions are automatically logged
- **Log Details**: User, action, resource, timestamp, IP address, status
- **Log Retrieval**: Query audit logs with filters
- **Compliance Support**: Detailed logs support regulatory compliance

---

## 5. How to Use the System / Component

### 5.1 Getting Started

#### Step 1: Obtain Access Credentials
1. Contact your system administrator to receive username and password
2. Default admin credentials (if applicable):
   - Username: `admin`
   - Password: `admin123`
   - **Important**: Change default password immediately

#### Step 2: Authenticate and Get Token
1. Send a POST request to `/api/v1/auth/login` with your credentials:
   ```json
   {
     "username": "your_username",
     "password": "your_password"
   }
   ```
2. Receive access token and refresh token in response
3. Store the access token securely (it expires in 15 minutes)
4. Use the refresh token to obtain new access tokens when needed

#### Step 3: Use Token for All Requests
- Include the token in the `Authorization` header of all subsequent requests:
  ```
  Authorization: Bearer <your_access_token>
  ```

### 5.2 Common Usage Scenarios

#### Scenario 1: Monitor Building Energy Consumption

**Goal**: View current energy consumption and device status for a building

**Steps**:
1. **Get Building Dashboard**:
   - Send GET request to `/api/v1/analytics/dashboards/building/{buildingId}`
   - Include your access token in the Authorization header
   - Response includes: device count, online devices, active anomalies, KPIs, forecast summary

2. **View Device Status**:
   - Send GET request to `/api/v1/iot/devices?buildingId={buildingId}`
   - Filter by status: `?status=ONLINE` to see only online devices
   - Response includes: device list with status, last seen, location

3. **Check for Anomalies**:
   - Send GET request to `/api/v1/analytics/anomalies?buildingId={buildingId}&severity=HIGH`
   - Review detected anomalies
   - Acknowledge important anomalies using POST `/api/v1/analytics/anomalies/acknowledge`

**Expected Outputs**:
- Dashboard data showing current consumption, device status, and KPIs
- List of devices with their current status
- List of detected anomalies with severity and details

#### Scenario 2: Generate Energy Consumption Report

**Goal**: Create a monthly energy consumption report for analysis

**Steps**:
1. **Request Report Generation**:
   - Send POST request to `/api/v1/analytics/reports/generate`
   - Include request body:
     ```json
     {
       "buildingId": "building-001",
       "type": "ENERGY_CONSUMPTION",
       "from": "2024-01-01T00:00:00Z",
       "to": "2024-01-31T23:59:59Z",
       "options": {
         "includeBreakdown": true,
         "includeRecommendations": true,
         "compareWithPreviousPeriod": true
       }
     }
     ```
   - Response includes report ID and status (PENDING)

2. **Wait for Report Generation**:
   - Report generation is asynchronous
   - Poll the report status or wait a few minutes
   - Check status using GET `/api/v1/analytics/reports/{reportId}`

3. **Retrieve Completed Report**:
   - Once status is COMPLETED, retrieve the full report
   - Report includes: summary, breakdown by device type, recommendations, comparisons

**Expected Outputs**:
- Report with total consumption, peak demand, average daily consumption
- Breakdown by device category (HVAC, lighting, equipment)
- Recommendations for energy savings
- Comparison with previous period (if requested)

#### Scenario 3: Control a Device (Adjust Temperature)

**Goal**: Change the temperature setpoint of an HVAC unit

**Steps**:
1. **Verify Device Status**:
   - Send GET request to `/api/v1/iot/devices/{deviceId}`
   - Ensure device status is ONLINE

2. **Send Command**:
   - Send POST request to `/api/v1/iot/device-control/{deviceId}/command`
   - Include command details:
     ```json
     {
       "command": "SET_TEMPERATURE",
       "params": {
         "temperature": 24.0,
         "unit": "celsius"
       }
     }
     ```
   - Response includes command ID and status (SENT)

3. **Verify Command Execution**:
   - Wait a few seconds for device acknowledgment
   - Check command status using GET `/api/v1/iot/device-control/{deviceId}/commands`
   - Status should change to APPLIED when device confirms execution

**Expected Outputs**:
- Command created with unique command ID
- Command status updates: PENDING → SENT → APPLIED
- Device temperature changes to the requested value

#### Scenario 4: Generate Forecast and Apply Optimization

**Goal**: Predict energy demand and automatically optimize consumption

**Steps**:
1. **Generate Forecast**:
   - Send POST request to `/api/v1/forecast/generate`
   - Include request:
     ```json
     {
       "buildingId": "building-001",
       "type": "DEMAND",
       "horizonHours": 24,
       "includeWeather": true,
       "includeTariffs": true
     }
     ```
   - Response includes forecast ID

2. **Check Peak Load**:
   - Send POST request to `/api/v1/forecast/peak-load`
   - Review predicted peak time and value
   - Check if threshold is exceeded

3. **Generate Optimization Scenario**:
   - Send POST request to `/api/v1/optimization/generate`
   - Specify optimization type (e.g., PEAK_SHAVING):
     ```json
     {
       "buildingId": "building-001",
       "name": "Peak Reduction",
       "type": "PEAK_SHAVING",
       "forecastId": "<forecast_id_from_step_1>",
       "useTariffData": true,
       "constraints": {
         "minTemperature": 20.0,
         "maxTemperature": 26.0,
         "preserveComfort": true
       }
     }
     ```
   - Review expected savings in response

4. **Apply Optimization** (if approved):
   - Send POST request to `/api/v1/optimization/send-to-iot`
   - Include scenario ID and execution preference
   - System automatically executes actions on devices

5. **Monitor Execution**:
   - Check status using GET `/api/v1/iot/optimization/status/{scenarioId}`
   - View progress and individual action status

**Expected Outputs**:
- Forecast with predictions for next 24 hours
- Peak load prediction with recommendations
- Optimization scenario with expected savings
- Execution progress and results

#### Scenario 5: Register a New Device

**Goal**: Add a new IoT device to the system

**Steps**:
1. **Register Device**:
   - Send POST request to `/api/v1/iot/devices/register`
   - Include device information:
     ```json
     {
       "deviceId": "hvac-003",
       "type": "HVAC",
       "model": "Carrier 24ACC636A003",
       "location": {
         "buildingId": "building-001",
         "floor": "3",
         "room": "Office 301"
       },
       "capabilities": ["SET_TEMPERATURE", "SET_MODE", "TURN_OFF"],
       "metadata": {
         "firmwareVersion": "2.1.3",
         "manufacturer": "Carrier"
       }
     }
     ```
   - Response confirms device registration

2. **Verify Registration**:
   - Retrieve device using GET `/api/v1/iot/devices/{deviceId}`
   - Confirm all information is correct

**Expected Outputs**:
- Device registered with unique ID
- Device appears in device listings
- Device can now receive telemetry and commands

### 5.3 API Usage Patterns

#### Authentication Pattern
```
1. Login → Get tokens
2. Use access token for all requests
3. When token expires → Refresh token
4. Logout when done
```

#### Data Retrieval Pattern
```
1. List resources (with filters)
2. Get specific resource by ID
3. Apply pagination for large datasets
```

#### Command Execution Pattern
```
1. Verify device status
2. Send command
3. Monitor command status
4. Verify execution
```

#### Report Generation Pattern
```
1. Request report generation
2. Wait for completion (poll status)
3. Retrieve completed report
```

### 5.4 Inter-Service Communication (For System Integrators)

If you are integrating EMSIB with other systems:

#### Token Validation
- Other services validate tokens by calling Security Service
- Endpoint: `GET /api/v1/auth/validate-token`
- Include token in Authorization header

#### Permission Checking
- Check user permissions before allowing actions
- Endpoint: `POST /api/v1/auth/check-permissions`
- Provide userId, resource, and action

#### Audit Logging
- Log significant actions for compliance
- Endpoint: `POST /api/v1/audit/log`
- Include user, action, resource, and status information

---

## 6. Error Handling & System Behavior

### 6.1 Common Error Types

#### Authentication Errors

**401 Unauthorized**
- **Cause**: Missing, invalid, or expired access token
- **What happens**: Request is rejected, no data is returned
- **What to do**: 
  - Check that Authorization header is included
  - Verify token is correct and not expired
  - Refresh token using `/api/v1/auth/refresh` endpoint
  - Re-authenticate if refresh token is also expired

**403 Forbidden**
- **Cause**: User lacks required permissions for the requested action
- **What happens**: Request is rejected with permission denied message
- **What to do**: 
  - Verify user has appropriate role/permissions
  - Contact administrator to grant required permissions
  - Use a different account with sufficient privileges

#### Validation Errors

**400 Bad Request**
- **Cause**: Invalid request data (missing required fields, wrong data types, invalid values)
- **What happens**: Request is rejected, error details are returned
- **What to do**: 
  - Review error message for specific validation issues
  - Check request body format and required fields
  - Correct invalid values and resubmit

**Example Error Response**:
```json
{
  "error": "Validation failed",
  "details": {
    "field": "buildingId",
    "message": "buildingId is required"
  }
}
```

#### Resource Not Found Errors

**404 Not Found**
- **Cause**: Requested resource (device, building, report, etc.) does not exist
- **What happens**: Request is rejected, resource not found message returned
- **What to do**: 
  - Verify resource ID is correct
  - Check if resource was deleted
  - List resources to find correct ID

#### Server Errors

**500 Internal Server Error**
- **Cause**: Unexpected server error (database issues, service failures, etc.)
- **What happens**: Request fails, error logged on server
- **What to do**: 
  - Retry request after a short delay
  - If persistent, contact system administrator
  - Check system health endpoints

**503 Service Unavailable**
- **Cause**: Service is temporarily unavailable (maintenance, overload, dependency failure)
- **What happens**: Request cannot be processed
- **What to do**: 
  - Wait and retry after a few minutes
  - Check service health endpoint: `GET /health`
  - Contact administrator if issue persists

### 6.2 Error Response Format

All errors follow a consistent format:

```json
{
  "error": "Error type or message",
  "message": "Detailed error description",
  "details": {
    "field": "specific field with issue",
    "value": "invalid value provided"
  },
  "timestamp": "2024-01-20T10:30:00Z",
  "path": "/api/v1/iot/devices"
}
```

### 6.3 System Behavior in Error Situations

#### Partial Failures

**Bulk Operations**
- If some items in a bulk operation fail, the system processes successful items
- Response indicates how many succeeded and how many failed
- Example: Bulk telemetry ingestion may succeed for 95 out of 100 items

**Optimization Scenarios**
- If some actions in an optimization scenario fail, others continue
- Progress tracking shows which actions succeeded and which failed
- System attempts to complete as many actions as possible

#### Timeout Behavior

**Command Timeouts**
- Commands sent to devices timeout after 30 seconds if no acknowledgment
- Command status changes to TIMEOUT
- **What to do**: 
  - Check device connectivity
  - Verify device is online
  - Retry command if appropriate

**Request Timeouts**
- Long-running operations (report generation, forecast generation) are asynchronous
- System returns immediately with a status of PENDING
- Users must poll for completion or wait for webhook notification (if configured)

#### Data Consistency

**Concurrent Updates**
- If multiple users modify the same resource simultaneously, last write wins
- System does not lock resources during updates
- **Recommendation**: Coordinate updates or use version numbers if available

**Eventual Consistency**
- Some data (dashboards, KPIs) may have slight delays due to aggregation
- Data is eventually consistent across services
- **Recommendation**: Allow a few seconds for data to propagate

### 6.4 What Users Should Do When Errors Occur

#### Immediate Actions

1. **Read the Error Message**: Error responses include detailed information about what went wrong
2. **Check Input Data**: Verify all required fields are provided and values are valid
3. **Verify Permissions**: Ensure your account has necessary permissions
4. **Check Token Validity**: Ensure access token is not expired

#### Retry Strategies

1. **Transient Errors** (500, 503): Wait 5-10 seconds and retry
2. **Rate Limiting**: If rate limited, wait before retrying
3. **Network Issues**: Check network connectivity and retry

#### Escalation

1. **Persistent Errors**: If errors persist after retries, contact system administrator
2. **Data Issues**: If data appears incorrect, report to administrator
3. **Service Outages**: If service is completely unavailable, contact operations team

#### Error Logging

- System automatically logs all errors for troubleshooting
- Administrators can review error logs to identify patterns
- Users should note error timestamps and request details when reporting issues

---

## 7. Best Practices & Recommendations

### 7.1 Authentication and Security

#### Token Management
- **Store tokens securely**: Never expose tokens in logs, URLs, or client-side code
- **Refresh proactively**: Refresh tokens before they expire to avoid interruptions
- **Use HTTPS**: Always use HTTPS in production to protect tokens in transit
- **Logout properly**: Call logout endpoint when user session ends

#### Password Security
- **Use strong passwords**: Minimum 8 characters, mix of letters, numbers, and symbols
- **Change default passwords**: Immediately change any default passwords
- **Don't share credentials**: Each user should have their own account

### 7.2 Efficient Data Usage

#### Pagination
- **Use pagination**: Always use pagination for list endpoints to avoid large responses
- **Reasonable page sizes**: Use page sizes between 10-100 items
- **Example**: `GET /api/v1/iot/devices?page=1&limit=20`

#### Filtering
- **Apply filters**: Use query parameters to filter results before retrieval
- **Reduce data transfer**: Only request data you need
- **Example**: `GET /api/v1/iot/devices?buildingId=building-001&status=ONLINE`

#### Caching
- **Cache static data**: Cache device lists, user information, etc.
- **Respect cache headers**: Follow cache-control headers if provided
- **Refresh periodically**: Update cached data periodically

### 7.3 Device Management

#### Device Registration
- **Use descriptive IDs**: Use meaningful device IDs (e.g., `hvac-floor3-room301`)
- **Complete information**: Provide all available device information during registration
- **Verify capabilities**: Accurately list device capabilities to enable proper control

#### Device Monitoring
- **Regular status checks**: Periodically check device status to identify offline devices
- **Monitor last seen**: Track `lastSeen` timestamps to identify connectivity issues
- **Set up alerts**: Configure alerts for critical devices going offline

### 7.4 Telemetry Data

#### Data Quality
- **Send accurate timestamps**: Include accurate timestamps in telemetry data
- **Consistent units**: Use consistent units (e.g., always Celsius for temperature)
- **Complete metrics**: Include all relevant metrics in telemetry readings

#### Data Frequency
- **Appropriate frequency**: Send telemetry at appropriate intervals (not too frequent, not too sparse)
- **Bulk ingestion**: Use bulk endpoints for multiple readings to improve efficiency
- **MQTT for real-time**: Use MQTT for high-frequency real-time data

### 7.5 Forecasting and Optimization

#### Forecast Accuracy
- **Sufficient historical data**: Ensure adequate historical data exists before generating forecasts
- **Include weather data**: Enable weather integration for better accuracy
- **Appropriate horizons**: Use appropriate forecast horizons (24-48 hours typically best)

#### Optimization Scenarios
- **Review before execution**: Always review optimization scenarios before applying
- **Test with dry run**: Use dry run mode to test scenarios without execution
- **Set constraints**: Define appropriate constraints (temperature limits, comfort requirements)
- **Monitor execution**: Monitor scenario execution progress and results

### 7.6 Reporting and Analytics

#### Report Generation
- **Reasonable time ranges**: Use appropriate time ranges (not too long, not too short)
- **Include options**: Use report options to get relevant breakdowns and comparisons
- **Schedule reports**: Generate reports during off-peak hours for large datasets

#### Dashboard Usage
- **Refresh appropriately**: Don't refresh dashboards too frequently (every 30-60 seconds is sufficient)
- **Use building-specific dashboards**: Use building dashboards for focused analysis
- **Combine with reports**: Use dashboards for quick views, reports for detailed analysis

#### Anomaly Management
- **Regular review**: Regularly review detected anomalies
- **Acknowledge promptly**: Acknowledge important anomalies promptly
- **Investigate high severity**: Prioritize investigation of HIGH and CRITICAL anomalies
- **Mark false positives**: Mark false positives to improve detection accuracy

### 7.7 Tips to Avoid Common Mistakes

#### Common Mistakes to Avoid

1. **Expired Tokens**: Not refreshing tokens before expiration
   - **Solution**: Implement token refresh logic or check token expiry

2. **Missing Required Fields**: Sending requests without required fields
   - **Solution**: Always check API documentation for required fields

3. **Wrong Data Types**: Sending strings instead of numbers or vice versa
   - **Solution**: Validate data types before sending requests

4. **Invalid Device IDs**: Using device IDs that don't exist
   - **Solution**: List devices first to get correct IDs

5. **Ignoring Errors**: Not handling error responses properly
   - **Solution**: Always check response status and handle errors appropriately

6. **Too Frequent Requests**: Making requests too frequently
   - **Solution**: Implement rate limiting or caching on client side

7. **Not Verifying Commands**: Not checking if commands were actually applied
   - **Solution**: Always verify command status after sending

8. **Large Time Ranges**: Requesting reports or data for very long time ranges
   - **Solution**: Break into smaller time ranges or use appropriate aggregation

### 7.8 Performance Optimization

#### Request Optimization
- **Batch operations**: Use bulk endpoints when possible
- **Parallel requests**: Make independent requests in parallel (with rate limiting)
- **Minimize round trips**: Combine related data requests when possible

#### Data Optimization
- **Use aggregations**: Use time-series aggregations instead of raw data when appropriate
- **Selective fields**: Request only needed fields if API supports field selection
- **Compression**: Enable HTTP compression if supported

---

## 8. Limitations

### 8.1 Known Constraints

#### Forecast Limitations
- **Maximum Horizon**: Forecasts are limited to 168 hours (7 days) maximum
- **Historical Data Requirement**: Accurate forecasts require sufficient historical data (minimum 30 days recommended)
- **External Dependencies**: Weather and tariff data availability affects forecast quality
- **ML Model Dependency**: Advanced predictions require external ML service availability

#### Report Generation
- **Asynchronous Processing**: Reports are generated asynchronously; users must wait or poll for completion
- **Large Time Ranges**: Very large time ranges may take significant time to process
- **Data Retention**: Reports are retained for 90 days by default, then automatically deleted

#### Anomaly Detection
- **Basic Detection**: Current implementation uses threshold-based detection; no advanced ML-based detection
- **False Positives**: May generate false positives that need manual review
- **Detection Delay**: Anomalies are detected after data is received, not in real-time

#### Time-Series Queries
- **Minimum Interval**: Minimum aggregation interval is 1 minute
- **Limited Intervals**: Only standard intervals (1m, 5m, 15m, 1h, 1d) are supported
- **Query Performance**: Very large time ranges may be slow

#### Device Management
- **Command Timeout**: Commands timeout after 30 seconds if device doesn't respond
- **No Device Authentication**: Devices authenticate via MQTT broker, not per-device tokens
- **Single MQTT Broker**: No MQTT broker clustering support
- **No Batch Commands**: Commands are sent individually, not batched

#### Data Retention
- **Telemetry Retention**: No automatic telemetry data cleanup (implement TTL indexes manually)
- **Command History**: Command history is retained indefinitely
- **Audit Logs**: Audit logs are retained based on system configuration

### 8.2 Assumptions Made by the System

#### Data Availability
- **Continuous Data**: System assumes devices send telemetry data regularly
- **Historical Data**: Forecasting assumes sufficient historical data exists
- **External Services**: System assumes external services (weather, tariffs) are available

#### Device Behavior
- **MQTT Connectivity**: System assumes devices maintain MQTT connectivity
- **Command Acknowledgment**: System assumes devices acknowledge commands within timeout period
- **Device Capabilities**: System assumes device capabilities are accurately registered

#### User Behavior
- **Token Management**: System assumes users properly manage and refresh tokens
- **Permission Awareness**: System assumes users understand their permissions
- **Error Handling**: System assumes users handle errors appropriately

### 8.3 Situations Where Functionality May Be Limited

#### Service Dependencies
- **Security Service Down**: If Security Service is unavailable, all authentication fails
- **IoT Service Down**: Device control and telemetry ingestion unavailable
- **Forecast Service Down**: Forecasting and optimization unavailable
- **Analytics Service Down**: Reports, dashboards, and analytics unavailable

#### Database Issues
- **MongoDB Unavailable**: All data operations fail if database is unavailable
- **Database Performance**: Slow database performance affects all operations

#### Network Issues
- **MQTT Broker Unavailable**: Real-time device communication fails
- **External API Failures**: Weather, tariff, or ML service failures affect related features
- **Network Latency**: High latency affects response times

#### Data Quality Issues
- **Insufficient Historical Data**: Forecasting accuracy decreases with limited data
- **Inconsistent Telemetry**: Missing or irregular telemetry affects analytics
- **Device Registration Errors**: Incorrect device information affects control capabilities

#### Scale Limitations
- **Large Number of Devices**: Very large numbers of devices (thousands) may affect performance
- **High Telemetry Volume**: Extremely high telemetry frequency may overwhelm system
- **Concurrent Users**: Very high concurrent user load may affect response times

### 8.4 Workarounds and Alternatives

#### When Forecasts Are Unavailable
- Use historical averages as estimates
- Manually analyze historical patterns
- Rely on real-time monitoring instead

#### When Reports Are Slow
- Use smaller time ranges
- Generate reports during off-peak hours
- Use dashboards for quick insights instead

#### When Devices Are Offline
- Check device connectivity
- Verify MQTT broker status
- Use HTTP endpoints as fallback if device supports it

#### When Services Are Down
- Contact system administrator
- Use cached data if available
- Wait for service restoration

---

## 9. Conclusion

### What Users Can Achieve

By using the EMSIB platform, users can:

1. **Gain Complete Visibility**: Monitor all building systems and devices in real-time, understanding energy consumption patterns, device health, and operational status

2. **Reduce Energy Costs**: Through automated optimization, peak shaving, and intelligent forecasting, users can significantly reduce energy consumption and operational costs

3. **Improve Operational Efficiency**: Automated device control, anomaly detection, and optimization scenarios reduce manual intervention and improve building operations

4. **Make Data-Driven Decisions**: Comprehensive analytics, reports, and KPIs provide the insights needed to make informed decisions about building operations and energy management

5. **Prevent Issues**: Early anomaly detection and predictive forecasting help identify and address problems before they cause significant impact

6. **Ensure Compliance**: Detailed audit logs and reporting capabilities support regulatory compliance and accountability

### Why It Is Useful Within the Overall Architecture

EMSIB serves as a critical component in modern smart building infrastructure:

- **Centralized Management**: Provides a single platform for managing diverse IoT devices and building systems
- **Intelligent Automation**: Enables automated optimization and control, reducing manual effort
- **Data Integration**: Integrates data from multiple sources (devices, weather, tariffs) to provide comprehensive insights
- **Scalable Architecture**: Microservices architecture allows the system to scale and evolve with changing needs
- **Interoperability**: RESTful APIs and MQTT support enable integration with existing building management systems
- **Security**: Centralized authentication and authorization ensure secure access to building systems

The platform transforms building management from reactive to proactive, enabling organizations to optimize energy consumption, reduce costs, and improve operational efficiency while maintaining comfort and meeting sustainability goals.

---

## Appendix A: Quick Reference

### Base URLs

| Service | URL |
|---------|-----|
| Security Service | `http://localhost:8080/api/v1` |
| Forecast Service | `http://localhost:8082/api/v1` |
| IoT Control Service | `http://localhost:8083/api/v1` |
| Analytics Service | `http://localhost:8084/api/v1` |

### Common Endpoints

| Purpose | Method | Endpoint |
|---------|--------|----------|
| Login | POST | `/api/v1/auth/login` |
| Get Devices | GET | `/api/v1/iot/devices` |
| Send Command | POST | `/api/v1/iot/device-control/{deviceId}/command` |
| Generate Forecast | POST | `/api/v1/forecast/generate` |
| Get Dashboard | GET | `/api/v1/analytics/dashboards/building/{buildingId}` |
| Generate Report | POST | `/api/v1/analytics/reports/generate` |
| List Anomalies | GET | `/api/v1/analytics/anomalies` |

### Default Roles

| Role | Description |
|------|-------------|
| `admin` | Full system access |
| `user` | Basic read access |
| `building_manager` | Building and device management |
| `energy_analyst` | Read-only analytics access |

---

## Appendix B: Support and Contact

For technical support, issues, or questions:

1. **Check Documentation**: Review this user documentation and service-specific README files
2. **Review Error Messages**: Error responses include detailed information
3. **Contact Administrator**: Reach out to your system administrator for account or permission issues
4. **System Health**: Check `/health` endpoints to verify service status

---

**End of User Documentation**
