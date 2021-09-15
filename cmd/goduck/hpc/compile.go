package hpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	versionRegexp = regexp.MustCompile(`[0-9]+\\.[0-9]+\\.[0-9]+`)
	legacyRegexp  = regexp.MustCompile(`0\\.(9\\..*|1\\.[01])`)
	paramsLegacy  = []string{
		"--binary",       // Request to output the contract in binary (hexadecimal).
		"file",           //
		"--json-abi",     // Request to output the contract's JSON ABI interface.
		"file",           //
		"--natspec-user", // Request to output the contract's Natspec user documentation.
		"file",           //
		"--natspec-dev",  // Request to output the contract's Natspec developer documentation.
		"file",
	}
	paramsNew = []string{
		"--bin",      // Request to output the contract in binary (hexadecimal).
		"--abi",      // Request to output the contract's JSON ABI interface.
		"--userdoc",  // Request to output the contract's Natspec user documentation.
		"--devdoc",   // Request to output the contract's Natspec developer documentation.
		"--optimize", // code optimizer switched on
		"--evm-version",
		"homestead",
		"-o", // output directory
	}
)

type Contract struct {
	Code string       `json:"code"`
	Info ContractInfo `json:"info"`
}

type ContractInfo struct {
	Source          string      `json:"source"`
	Language        string      `json:"language"`
	LanguageVersion string      `json:"languageVersion"`
	CompilerVersion string      `json:"compilerVersion"`
	CompilerOptions string      `json:"compilerOptions"`
	AbiDefinition   interface{} `json:"abiDefinition"`
	UserDoc         interface{} `json:"userDoc"`
	DeveloperDoc    interface{} `json:"developerDoc"`
}

type Solidity struct {
	solcPath    string
	version     string
	fullVersion string
	legacy      bool
	isSolcjs    bool
}

func NewCompiler(solcPath string) (sol *Solidity, err error) {
	// set default solc
	if len(solcPath) == 0 {
		solcPath = "solc"
	}
	solcPath, err = exec.LookPath(solcPath)
	if err != nil {
		solcPath = "solcjs"
		if solcPath, err = exec.LookPath(solcPath); err != nil {
			return nil, &exec.Error{Name: "solc and solcjs", Err: exec.ErrNotFound}
		}
	}

	cmd := exec.Command(solcPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return
	}

	fullVersion := out.String()
	version := versionRegexp.FindString(fullVersion)
	legacy := legacyRegexp.MatchString(version)

	sol = &Solidity{
		solcPath:    solcPath,
		version:     version,
		fullVersion: fullVersion,
		legacy:      legacy,
	}

	if strings.HasSuffix(sol.solcPath, "solcjs") {
		sol.isSolcjs = true
	}

	return
}

func (sol *Solidity) Info() string {
	return fmt.Sprintf("%s\npath: %s", sol.fullVersion, sol.solcPath)
}

func (sol *Solidity) Version() string {
	return sol.version
}

// Compile builds and returns all the contracts contained within a source string.
func (sol *Solidity) Compile(source string) (map[string]*Contract, error) {
	// Short circuit if no source code was specified
	if len(source) == 0 {
		return nil, errors.New("solc: empty source string")
	}
	// Create a safe place to dump compilation output
	wd, err := ioutil.TempDir("", "solc")
	if err != nil {
		return nil, fmt.Errorf("solc: failed to create temporary build folder: %v", err)
	}
	defer os.RemoveAll(wd)

	// Assemble the compiler command, change to the temp folder and capture any errors
	stderr := new(bytes.Buffer)

	var params []string
	if sol.legacy {
		params = paramsLegacy
	} else {
		params = paramsNew
		params = append(params, wd)
	}
	compilerOptions := strings.Join(params, " ")

	cmd := exec.Command(sol.solcPath, params...)
	cmd.Stderr = stderr
	if !sol.isSolcjs {
		cmd.Stdin = strings.NewReader(source)

	} else {
		sourceFile, err := ioutil.TempFile(wd, "source")
		if err != nil {
			return nil, err
		}

		_, err = sourceFile.Write([]byte(source))
		if err != nil {
			return nil, err
		}

		cmd.Dir = wd
		cmd.Args = append(cmd.Args, path.Base(sourceFile.Name()))
	}

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("solc: %v\n%s", err, string(stderr.Bytes()))
	}
	// Sanity check that something was actually built
	matches, _ := filepath.Glob(filepath.Join(wd, "*.bin*"))
	if len(matches) < 1 {
		return nil, fmt.Errorf("solc: no build results found")
	}
	// Compilation succeeded, assemble and return the contracts
	contracts := make(map[string]*Contract)
	for _, path := range matches {
		_, file := filepath.Split(path)
		base := strings.Split(file, ".")[0]

		// Parse the individual compilation results (code binary, ABI definitions, user and dev docs)
		var binary []byte
		binext := ".bin"
		if sol.legacy {
			binext = ".binary"
		}
		if binary, err = ioutil.ReadFile(filepath.Join(wd, base+binext)); err != nil {
			return nil, fmt.Errorf("solc: error reading compiler output for code: %v", err)
		}

		var abi interface{}
		if blob, err := ioutil.ReadFile(filepath.Join(wd, base+".abi")); err != nil {
			return nil, fmt.Errorf("solc: error reading abi definition: %v", err)
		} else if err = json.Unmarshal(blob, &abi); err != nil {
			return nil, fmt.Errorf("solc: error parsing abi definition: %v", err)
		}

		var userdoc interface{}
		var devdoc interface{}

		if !sol.isSolcjs {
			if blob, err := ioutil.ReadFile(filepath.Join(wd, base+".docuser")); err != nil {
				return nil, fmt.Errorf("solc: error reading user doc: %v", err)
			} else if err = json.Unmarshal(blob, &userdoc); err != nil {
				return nil, fmt.Errorf("solc: error parsing user doc: %v", err)
			}

			if blob, err := ioutil.ReadFile(filepath.Join(wd, base+".docdev")); err != nil {
				return nil, fmt.Errorf("solc: error reading dev doc: %v", err)
			} else if err = json.Unmarshal(blob, &devdoc); err != nil {
				return nil, fmt.Errorf("solc: error parsing dev doc: %v", err)
			}
		}

		// Assemble the final contract
		contracts[base] = &Contract{
			Code: "0x" + string(binary),
			Info: ContractInfo{
				Source:          source,
				Language:        "Solidity",
				LanguageVersion: sol.version,
				CompilerVersion: sol.version,
				CompilerOptions: compilerOptions,
				AbiDefinition:   abi,
				UserDoc:         userdoc,
				DeveloperDoc:    devdoc,
			},
		}
	}
	return contracts, nil
}
