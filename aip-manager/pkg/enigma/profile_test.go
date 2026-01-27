package enigma

import (
	"path/filepath"
	"strings"
	"testing"
)

const sampleProfile = `
<Root>
  <Input>
    <FileName>E:\Enigma\files\scr.exe</FileName>
    <ProductName/>
  </Input>
  <Output>
    <SameAsInputFile>False</SameAsInputFile>
    <FileName>E:\Enigma\packed-files\scr.exe</FileName>
  </Output>
  <AdvanceInput>
    <ProtectAfter>True</ProtectAfter>
    <Files Count="4">
      <File>
        <Input>E:\Enigma\files\CyberCommonRules.dll</Input>
        <Output>E:\Enigma\packed-files\CyberCommonRules.dll</Output>
      </File>
      <File>
        <Input>E:\Enigma\files\CyberExpert.dll</Input>
        <Output>E:\Enigma\packed-files\CyberExpert.dll</Output>
      </File>
    </Files>
  </AdvanceInput>
</Root>`

func TestRewriteProfileXML_MapsPathsToDirs(t *testing.T) {
	unpacked := filepath.Join("X:", "ws", "installer_src", "files", "Unpacked")
	packed := filepath.Join("X:", "ws", "installer_src", "files", "Packed")

	out, err := rewriteEnigmaProfileXML([]byte(sampleProfile), unpacked, packed)
	if err != nil {
		t.Fatalf("rewrite failed: %v", err)
	}
	s := string(out)

	expect := []string{
		filepath.Join(unpacked, "scr.exe"),
		filepath.Join(packed, "scr.exe"),
		filepath.Join(unpacked, "CyberCommonRules.dll"),
		filepath.Join(packed, "CyberCommonRules.dll"),
		filepath.Join(unpacked, "CyberExpert.dll"),
		filepath.Join(packed, "CyberExpert.dll"),
	}
	for _, want := range expect {
		if !strings.Contains(s, want) {
			t.Fatalf("rewritten profile missing %q\n--- got ---\n%s", want, s)
		}
	}
}
