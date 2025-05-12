# calculator_go_5: Web-service "Calculator" for Yandex Lyceum (5 Sprint)

### Description
This project implements a web service that evaluates arithmetic expressions submitted by the user via an HTTP request.

# Starting the service
  1. Check that you have [Go](https://go.dev/dl/) and [gcc]() compiler installed
  2. Clone the project from GitHub:
     ```
      git clone https://github.com/arniknz/calculator_go_5.git
     ```
  3. To start service
     ```
      cd .\calculator_go_5\
      go run cmd/orchestrator/main.go
      go run cmt/agent/main.go
     ```
  4. The service will be available at: ```http://localhost:8080/api/v1/calculate```

# Registration

Register:
```
    curl --location 'localhost:8080/api/v1/register' \
    --header 'Content-Type: application/json' \
    --data '{
      "login": "<your-login>",
      "password": "<your-password>"
    }'
```

Login:
```
    curl --location 'localhost:8080/api/v1/login' \
    --header 'Content-Type: application/json' \
    --data '{
      "login": "<your-login>",
      "password": "<your-password>"
    }'
```

You will get a jwt token:
```
    <your-mega-super-cool-jwt-token>
```

# Endpoints:

Note: If you have not registered/logined, you will receive a code: ```401 Unauthorized```

Add expresion calc:
```
    curl --location 'localhost:8080/api/v1/calculate' \
    --header 'Authorization: <your-mega-super-cool-jwt-token>'
    --header 'Content-Type: application/json' \
    --data '{
      "expression": "<your_expression>"
    }'
```

Status Codes: 201 - expression accepted, 422 - Invalid data, 500 - Something went wrond

Response Body

```
{
    "id": unique_id
}
```

Get list of expressions:
```curl --location 'localhost/api/v1/expressions'```
```
    curl --location 'localhost:8080/api/v1/expressions' \
    --header 'Authorization: <your-mega-super-cool-jwt-token>'
```

Response Body:
```
"expressions": [
        {
            "id": <id>,
            "status": <status: pending/progress/completed>,
            "result": <calculation result>
        },
        {
            "id": <id>,
            "status": <status: pending/progress/completed>,
            "result": <calculation result>
        }
    ]
```

Status Codes:
```
    200 - OK
    500 - Something went wrond
```

Get expression by id:
```curl --location 'localhost/api/v1/expressions/:id'```
```
    curl --location 'localhost:8080/api/v1/expressions/:id' \
    --header 'Authorization: <your-mega-super-cool-jwt-token>'
```

Status Codes:
```
    200 - OK
    404 - Expression not found
    500 - Something went wrond
```

# Console export variables
#### DEBUG for testing!
```
    export TIME_ADDITION_MS=50
    export TIME_SUBTRACTION_MS=50
    export TIME_MULTIPLICATIONS_MS=100
    export TIME_DIVISIONS_MS=100
    export COMPUTING_POWER=4
    export DEBUG=false
```


# Example requests
### cURL:
| METHOD | cURL request | Response | Status code |
| ------ | ------------ | -------- | ----------- |
| POST   |<code>curl --location 'http://localhost:8080/api/v1/calculate' --header 'Authorization: jwt-token' --header 'Content-Type: application/json' --data '{  "expression": "2+2*2"  }'</code>|<code>{"result": 6.000000}</code>| 200 |
| POST   |<code>curl --location 'http://localhost:8080/api/v1/calculate' --header 'Authorization: jwt-token' --header 'Content-Type: application/json' --data '{  "expression": "2 + "  }'</code>|<code>{"error": "Invalid expression"}</code>| 422 |
| GET    |<code>curl --request GET --url 'http://localhost:8080/api/v1/calculate' --header 'Authorization: jwt-token' --header 'Content-Type: application/json' --data '{  "expression": "2 + 1"  }'</code>|<code>{"error": "Only POST method is allowed"}</code>| 405 |
| POST   |<code>curl --location 'http://localhost:8080/api/v1/calculate' --header 'Authorization: jwt-token' --header 'Content-Type: application/json' --data '{  "bebebe": "2 + 2"  }'</code>|<code>{"error": "Bad request"}</code>| 400 |

### Internal Server Error: 500 status code
If an internal server error occurs, the service will return an error with status code 500
**<code>"error: Internal server error"</code>**

### Simplified explanation:
| METHOD | Json request | Response | Status code |
| ------ | ------------ | -------- | ----------- |
| POST   | <code>{  "expression": "2+2*2"  }</code>|<code>{"result": 6.000000}</code>| 200 |
| POST   | <code>{  "expression": "2 + "  }</code>|<code>{"error": "Invalid expression"}</code>| 422 |
| GET    | <code>{  "expression": "2 + 1"  }</code>|<code>{"error": "Only POST method is allowed"}</code>| 405 |
| POST   | <code>{  "bebebe": "2 + 2"  }</code>|<code>{"error": "Bad request"}</code>| 400 |

## Testing

### WARNING: TEST ONLY IN DEBUG MODE!

```
export DEBUG=true
```

```
cd test/
go test
```

After the test, the **calculator_5_go.db** database will appear in the current folder; for the next test, you must delete this database **EXACTLY** in the current folder.

Delete **test/calculator_5_go.db** after test

Do not delete **calculator_go_5/calculator_5_go.db**, because you may lose all your data!

# TUTORIAL: How to install gcc on [Windows](https://code.visualstudio.com/docs/cpp/config-mingw)
1. You can download the latest installer from the MSYS2 page or use [this direct link to the installer](https://github.com/msys2/msys2-installer/releases/download/2024-12-08/msys2-x86_64-20241208.exe).

2. Run the installer and follow the steps of the installation wizard. Note that MSYS2 requires 64 bit Windows 8.1 or newer.

3. In the wizard, choose your desired Installation Folder. Record this directory for later. In most cases, the recommended directory is acceptable. The same applies when you get to setting the start menu shortcuts step. When complete, ensure the Run MSYS2 now box is checked and select Finish. This will open a MSYS2 terminal window for you.

4. In this terminal, install the MinGW-w64 toolchain by running the following command:
    ```
    pacman -S --needed base-devel mingw-w64-ucrt-x86_64-toolchain
    ```

5. Accept the default number of packages in the ```toolchain``` group by pressing **Enter**.

6. Enter ```Y``` when prompted whether to proceed with the installation.

7. Add the path of your MinGW-w64 ```bin``` folder to the Windows ```PATH``` environment variable by using the following steps:

    1. In the Windows search bar, type **Settings** to open your **Windows Settings**.

    2. Search for **Edit environment variables for your account**.

    3. In your **User variables**, select the ```Path``` variable and then select **Edit**.

    4. Select **New** and add the MinGW-w64 destination folder you recorded during the installation process to the list. If you used the default settings above, then this will be the path: ```C:\msys64\ucrt64\bin```.

    5. Select **OK**, and then select **OK** again in the **Environment Variables** window to update the ```PATH``` environment variable. You have to reopen any console windows for the updated ```PATH``` environment variable to be available.

## Check your MinGW installation

To check that your MinGW-w64 tools are correctly installed and available, open a new Command Prompt and type:

```
gcc --version
g++ --version
gdb --version
```

You should see output that states which versions of GCC, g++ and GDB you have installed. If this is not the case: 
   1. Make sure your PATH variable entry matches the MinGW-w64 binary location where the toolchain was installed. If the compilers do not exist at that PATH entry, make sure you followed the previous instructions.

   2. If ```gcc``` has the correct output but not ```gdb```, then you need to install the packages you are missing from the MinGW-w64 toolset.
        - If on compilation you are getting the "The value of miDebuggerPath is invalid." message, one cause can be you are missing the ```mingw-w64-gdb``` package.

# TUTORIAL: How to install gcc on [Linux](https://askubuntu.com/questions/398489/how-to-install-build-essential)
```
sudo apt install build-essential
```
