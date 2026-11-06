//go:build unix

package terminal

import "github.com/cucumber/godog"

// registerCommandLineSteps registers the steps of features/command-line.feature.
func (j *journey) registerCommandLineSteps(sc *godog.ScenarioContext) {
	sc.Step(`^my data is in the folder "([^"]*)"$`, j.myDataIsInTheFolder)
	sc.Step(`^I ask the command line where my data is$`, j.iAskTheCommandLineWhereMyDataIs)
	sc.Step(`^it says my data is in the folder "([^"]*)"$`, j.itSaysMyDataIsInTheFolder)
	sc.Step(`^a server is running$`, j.aServerIsRunning)
	sc.Step(`^I ask the command line where my data is, connected to the server$`, j.iAskTheCommandLineWhereMyDataIsConnectedToTheServer)
	sc.Step(`^it says my data is on the server$`, j.itSaysMyDataIsOnTheServer)
	sc.Step(`^I give the command line the command "([^"]*)"$`, j.iGiveTheCommandLineTheCommand)
	sc.Step(`^I ask the command line to move the workspace "([^"]*)" to the "([^"]*)" plan$`, j.iAskTheCommandLineToMoveTheWorkspaceToThePlan)
	sc.Step(`^it fails, saying "([^"]*)"$`, j.itFailsSaying)
}

func (j *journey) myDataIsInTheFolder(folder string) error {
	j.dataDir = folder
	return nil
}

func (j *journey) iAskTheCommandLineWhereMyDataIs() error {
	return j.run("--data-dir", j.dataDir, "info")
}

func (j *journey) itSaysMyDataIsInTheFolder(folder string) error {
	if err := j.term.WaitFor("Data folder: "+folder, timeout); err != nil {
		return err
	}
	return j.succeeds()
}

func (j *journey) aServerIsRunning() error {
	return j.startServer()
}

func (j *journey) iAskTheCommandLineWhereMyDataIsConnectedToTheServer() error {
	return j.run("--server", j.serverURL, "info")
}

func (j *journey) itSaysMyDataIsOnTheServer() error {
	if err := j.term.WaitFor("Server: "+j.serverURL, timeout); err != nil {
		return err
	}
	return j.succeeds()
}

func (j *journey) iGiveTheCommandLineTheCommand(command string) error {
	return j.run(command)
}

func (j *journey) iAskTheCommandLineToMoveTheWorkspaceToThePlan(workspace, plan string) error {
	return j.run("admin", "set-plan", workspace, plan)
}

func (j *journey) itFailsSaying(message string) error {
	if err := j.term.WaitFor(message, timeout); err != nil {
		return err
	}
	return j.exitStatus(func(status int) bool { return status != 0 }, "a failure")
}
