# Define the path to vcvars64.bat
$vcvarsPath = "C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\VC\Auxiliary\Build\vcvars64.bat"

# Call vcvars64.bat and then run the Gradle command
cmd.exe /c "call `"$vcvarsPath`" && gradlew :build -Dquarkus.package.type=native -DskipTests"