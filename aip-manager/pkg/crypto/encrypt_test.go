// pkg/crypto/encrypt_test.go
package crypto

import (
	"path/filepath"
	"strings"
	"testing"

	"aip-manager/pkg/config"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	// This test ensures that what we encrypt can be successfully decrypted back to the original plaintext.
	testCases := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "Simple String",
			plaintext: "hello world",
		},
		{
			name:      "INI Modules Config",
			plaintext: "[Modules]\nEnable_Dac=false\n",
		},
		{
			name:      "INI STS Config",
			plaintext: "[general]\nlast_commit=abc-123\n[profiling]\nenable_profiling=0\n",
		},
		{
			name:      "Empty String",
			plaintext: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt the plaintext
			encryptedHex, err := Encrypt(tc.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() returned an unexpected error: %v", err)
			}

			// Decrypt the ciphertext
			decryptedText, err := Decrypt(encryptedHex, aesKeyString)
			if err != nil {
				t.Fatalf("Decrypt() returned an unexpected error: %v", err)
			}

			// Verify that the decrypted text matches the original plaintext
			if decryptedText != tc.plaintext {
				t.Errorf("Roundtrip failed. \nOriginal:  %q\nDecrypted: %q", tc.plaintext, decryptedText)
			}
		})
	}
}

func TestVectorValidation(t *testing.T) {
	// These test cases use the exact INI strings and the now-corrected ciphertexts to validate
	// that our Go implementation matches the verified C++ logic.
	// NOTE: The plaintext now uses '\n' (LF) instead of '\r\n' (CRLF) as per your finding.
	testCases := []struct {
		name        string
		iniString   string
		expectedHex string
	}{
		{
			name: "Modules INI Validation",
			iniString: "[Modules]\n" +
				"HPSP_remove_quarantine_file=false\n" +
				"AI_remove_quarantine_file=true\n" +
				"RakarL1_remove_quarantine_file=false\n" +
				"RakarL2_AM_remove_quarantine_file=true\n" +
				"RakarL3_remove_quarantine_file=false\n" +
				"HIDS_remove_quarantine_file=false\n" +
				"Firewall_remove_quarantine_file=false\n" +
				"Blacklist_remove_quarantine_file=true\n" +
				"Enable_Dac=false\n",
			expectedHex: "3562b3107539bae5e9ece48f9a0e9c7771b9c37fa1f2ba9de74bc706391851de36225afd0994c2608747879003e285e3e7820209c52541156b870c75489ef2c26f2598d2c7d8a9be282054378569baa2b94c1db3fdc0d334f62afdb267a606697c83cd2396b8b38f10997cc2be47c3846bed4608ae7aadd3c1e3cd3a7cfcb7b60e863ad12b7e8afe8f5781b56a58fa5e428d62b216a47558e9d85a162571fbfc109a72dcdc00f5017856410f9b876105b5adb09835d99b96bc8eaf00f4f2087baaf4c0c48e1c32f26c1c08ab2591079c0c5e95bec251f386e39d4487f39b5554bdbddfe7a3f6db57ab371433a905ef3376f8020fe05b901e406403617c8acf00504c17be773c3102237e26be3d25395f6dfbfbf4e14bec2f9508674a5e4537770291278a192f2d2633b7a10727f0598bad973cc099caba20e0ad194cc5240076",
		},
		{
			name: "STS INI Validation",
			iniString: "[profiling]\n" +
				"enable_profiling=0\n" +
				"[general]\n" +
				"hide_console=0\n" +
				"event_processor_workers=3\n" +
				"service_accept_value=0\n" +
				"disable_quarantine_file_content_for_ai_tests=1\n" +
				"do_not_remove_quarantined_files=0\n" +
				"enable_dllinjection=0\n" +
				"enable_network_driver=0\n" +
				"enable_self_defense=1\n" +
				"enable_antiransomware=0\n" +
				"quarantine_rakarl2am_malwares=1\n" +
				"quarantine_blacklist_malwares=1\n" +
				"quarantine_rakar_sigmaplus_malwares=0\n" +
				"quarantine_AI_malwares=1\n" +
				"quarantine_signature_malwares=0\n" +
				"last_commit=b75cd650adf06b34a5b09d038ac590b4e04a9bc2\n" +
				"[events]\n" +
				"process_EDRDriver_FileSystem_FileCreate=0\n" +
				"process_EDRDriver_FileSystem_FileOpen=0\n" +
				"process_EDRDriver_FileSystem_FileClose=1\n" +
				"process_EDRDriver_FileSystem_FileDelete=0\n" +
				"process_EDRDriver_FileSystem_FileRead=0\n" +
				"process_EDRDriver_FileSystem_FileWrite=0\n" +
				"process_EDRDriver_FileSystem_FileDataChange=0\n" +
				"process_EDRDriver_FileSystem_FileDataReadFull=0\n" +
				"process_EDRDriver_FileSystem_FileDataWriteFull=1\n" +
				"process_EDRDriver_FileSystem_AntiRansomware=0\n" +
				"process_EDRDriver_Network_Connect=0\n" +
				"process_EDRDriver_Process_CreateProcess=1\n" +
				"process_EDRDriver_Process_DeleteProcess=1\n" +
				"process_EDRDriver_Process_OpenProcess=1\n" +
				"process_EDRDriver_Registry_KeyNameChange=1\n" +
				"process_EDRDriver_Registry_CreateKey=1\n" +
				"process_EDRDriver_Registry_KeyValueSet=1\n" +
				"process_EDRDriver_Registry_DeleteKey=1\n" +
				"process_EDRDriver_Registry_DeleteValue=1\n" +
				"process_EDRDriver_Process_LoadImage=0\n" +
				"process_ETW_Network_ConnectionAttempted=0\n" +
				"[directories]\n" +
				"only_process_files_in_dir=\n",
			expectedHex: "a223c8b37136928ac1fcb2b9873a096140ff42526492d690f46a336731a0818ac84ac8165525a0ff9d551fa4381ae0ac580358e4a617449d192ce01f3ebb4ad291c2741c11679850cc512e548c5d9abeb15970685d86c4d71946823b36fdf3682a86c12a496be1134622c9c53bc01ca0025b1f905e8ac7a2d26e06b6f2ec837595b3997c43ab416fc7f36b8e09487ea8f4a0aef4fbaddc166eff09d0a603b4f516123e55112dd6d5ff190fd5fc381c1f68942de1126511277adcc1a3253674ca9ec87eef8e67f1ce7a438b97010ac6b2016e486faaca75c829f8ff4369efd088d5281d4d1724f9da334a6458ddee95197bab3eb90f3eb80cbe843cc3a35d05c37d55d8be42b84a725233b85760e40c1b4273b06cdf5ef996caeccef605a80cbd1406604b16d3b05dcf9d408a054ac7d98e475f185fd0a58fd1d084a0d38ba0b925f22681d98dfb91742cf26a45bc46297d337dbbe26afe28142b753a3b8c9da332cb2c6abe3f87e4123ca7c4733e3b2c8a9dcadb6da776668543fdb42b54e15712be4fc3fea81c9ea28349fad0698514fda1b1200974307d4055e63630896934eb744ca1d515ec3e0fbc6988cc6187c9614879e68994b56e2d172bbf0b11bb8f018a3c6b2a7f0a97febd2f406f554b05bbe1b834ab9991e3cff4a12552159377134855e49acf236871cff574339b057a2e0f9eb63084c99e317f3f0df7856fca3f23c63841fd24e22856b4b1b83cdabf907f99f4dcb2bceb71bc14021165b5a7b74776ea36572a2fd3903628afd9a158185c5c8303f1b07e8cad2418834d658c64a8127790643b0a470b9460d4db3fc15101395dffd5b22042e3cb09d46976a19c654ba8763e2d275becf4f1ad938ae400fd355a61c87e4fd72eacec188845c8e1d2d00a5849841027e28b37c78c2276dff6373b195524429932f9f645f00746a8b560e91c61ce8a48467f3550605e25588a5b15a4676f9852aaedde8d8b9183782e3d285416456690b10d1f89bca76ee2e9ebadc2ce51ccdfa8c4f1b5540fbbc35b7ab71e871a6e7c1578516c478ecaa692bbbc65b90f718efc1597ce2c0ed95c59c15a57fa75597ac1d4ef16b3f5d57824450821f59e17080fbcb99d744a0a35333e95cd0b032b7c1bdacf322760357748f8f47382bf0a3374cd76e08ad5321133568ebe254d91d066f627e1843c7a2b4bbde656614eb287d31d9c9b8a2a25b50143c8a16d3ee5ee8dca6863bd0f9849d385d45e2b69a7a0df86092ddb5fdc727efc92afb8cf2df44c189748a1b238f14d6f47c8f73c9ad43650465c8df44bb1a98871996a31ba7bf85476704eb9402e61e889c11ae77780ddd1c71c0ecba44c04f01eb232aa51f8fe323b6ad59396be70b8e2e093668b228113e0f62b4097f24a548fcab1f8aedfbb2663c21b6928f15f3f3f5fbed598ae0664e644f84fbf4f2face704a3640dd8a195446f30599f7e4285a38faecf83b4ea61910b0e6c552593b493942490902c4a94d73eab18562a85ebfdfc3ddc51f406d43b89333fc4e5ec25059be7b9f15324160c33579ece90d823b1159bcbec4792dce6bdf31959486e0a168e781631a5b27080b82946a420467d180e30baefe346a0fd6bd06bef91aba66216c3f0dcc709d5b235b62ee98be9370ca3edac7c3b6bb56436dd3b38baa8bf7acc79488467191e40f23f38c72350b9dd829f315c9031be4d653c39fa3b82e69c48f118979849f2f6014087c60e520437076761c9f8c588b791aab9dba197af2b6de5dca8afc1df4fb63400f1b97cc67dcef5b5cb751b892e7d53a302cf8036062b98eebfe6cf4b640ef23d0101ec80e40bb69981fad2f3abac95956a217c12e6ae3d02e2088f3b164d7ab45956b78d40b88d550f46be698db0a190de0a1cc3adec79a55b04eecb4d14509e96b1c1eddd13eadc3a67693f1ed82ebc7505e15d35e1a56a55832a74e421441ec0870199ee946e8e534a4e90bbe0a5d7f9961a8a340aadad63fae6d6d51134c8af",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt the known INI string
			encryptedHex, err := Encrypt(tc.iniString)
			if err != nil {
				t.Fatalf("Encrypt() returned an unexpected error: %v", err)
			}

			// Remove the "#x" prefix for comparison
			generatedHex := strings.TrimPrefix(encryptedHex, "#x")

			// Check if the output matches the known-good ciphertext
			if generatedHex != tc.expectedHex {
				t.Errorf("Ciphertext mismatch for %s.\nExpected: %s\nGot:      %s", tc.name, tc.expectedHex, generatedHex)
			}
		})
	}
}

// Test_Integration performs an integration test on the INI generation and encryption process.
// It loads a config from a file, generates the INI string (respecting key order),
// and then encrypts it, validating the final ciphertext against a known-good vector.
func Test_Integration(t *testing.T) {
	// 1. Specify the configuration from a dedicated test file.
	// This ensures the loader and the order-preserving logic are tested together.
	tests := []testConfig{
		{
			configPath: filepath.Join("testdata", "ai.yaml"),
			cases: []testCase{
				{
					name:        "AI Modules Encryption Validation",
					input:       "1",
					expectedHex: "3562b3107539bae5e9ece48f9a0e9c7771b9c37fa1f2ba9de74bc706391851de36225afd0994c2608747879003e285e3e7820209c52541156b870c75489ef2c26f2598d2c7d8a9be282054378569baa2b94c1db3fdc0d334f62afdb267a606697c83cd2396b8b38f10997cc2be47c3846bed4608ae7aadd3c1e3cd3a7cfcb7b60e863ad12b7e8afe8f5781b56a58fa5e428d62b216a47558e9d85a162571fbfc109a72dcdc00f5017856410f9b876105b5adb09835d99b96bc8eaf00f4f2087baaf4c0c48e1c32f26c1c08ab2591079c0c5e95bec251f386e39d4487f39b5554bdbddfe7a3f6db57ab371433a905ef3376f8020fe05b901e406403617c8acf00504c17be773c3102237e26be3d25395f6dfbfbf4e14bec2f9508674a5e4537770291278a192f2d2633b7a10727f0598bad973cc099caba20e0ad194cc5240076",
				},
				{
					name:        "AI STS Encryption Validation",
					input:       "2",
					expectedHex: "a223c8b37136928ac1fcb2b9873a096140ff42526492d690f46a336731a0818ac84ac8165525a0ff9d551fa4381ae0ac580358e4a617449d192ce01f3ebb4ad291c2741c11679850cc512e548c5d9abeb15970685d86c4d71946823b36fdf3682a86c12a496be1134622c9c53bc01ca0025b1f905e8ac7a2d26e06b6f2ec837595b3997c43ab416fc7f36b8e09487ea8f4a0aef4fbaddc166eff09d0a603b4f516123e55112dd6d5ff190fd5fc381c1f68942de1126511277adcc1a3253674ca9ec87eef8e67f1ce7a438b97010ac6b2016e486faaca75c829f8ff4369efd088d5281d4d1724f9da334a6458ddee95197bab3eb90f3eb80cbe843cc3a35d05c37d55d8be42b84a725233b85760e40c1b4273b06cdf5ef996caeccef605a80cbd1406604b16d3b05dcf9d408a054ac7d98e475f185fd0a58fd1d084a0d38ba0b925f22681d98dfb91742cf26a45bc46297d337dbbe26afe28142b753a3b8c9da332cb2c6abe3f87e4123ca7c4733e3b2c8a9dcadb6da776668543fdb42b54e15712be4fc3fea81c9ea28349fad0698514fda1b1200974307d4055e63630896934eb744ca1d515ec3e0fbc6988cc6187c9614879e68994b56e2d172bbf0b11bb8f018a3c6b2a7f0a97febd2f406f554b05bbe1b834ab9991e3cff4a12552159377134855e49acf236871cff574339b057a2e0f9eb63084c99e317f3f0df7856fca3f23c63841fd24e22856b4b1b83cdabf907f99f4dcb2bceb71bc14021165b5a7b74776ea36572a2fd3903628afd9a158185c5c8303f1b07e8cad2418834d658c64a8127790643b0a470b9460d4db3fc15101395dffd5b22042e3cb09d46976a19c654ba8763e2d275becf4f1ad938ae400fd355a61c87e4fd72eacec188845c8e1d2d00a5849841027e28b37c78c2276dff6373b195524429932f9f645f00746a8b560e91c61ce8a48467f3550605e25588a5b15a4676f9852aaedde8d8b9183782e3d285416456690b10d1f89bca76ee2e9ebadc2ce51ccdfa8c4f1b5540fbbc35b7ab71e871a6e7c1578516c478ecaa692bbbc65b90f718efc1597ce2c0ed95c59c15a57fa75597ac1d4ef16b3f5d57824450821f59e17080fbcb99d744a0a35333e95cd0b032b7c1bdacf322760357748f8f47382bf0a3374cd76e08ad5321133568ebe254d91d066f627e1843c7a2b4bbde656614eb287d31d9c9b8a2a25b50143c8a16d3ee5ee8dca6863bd0f9849d385d45e2b69a7a0df86092ddb5fdc727efc92afb8cf2df44c189748a1b238f14d6f47c8f73c9ad43650465c8df44bb1a98871996a31ba7bf85476704eb9402e61e889c11ae77780ddd1c71c0ecba44c04f01eb232aa51f8fe323b6ad59396be70b8e2e093668b228113e0f62b4097f24a548fcab1f8aedfbb2663c21b6928f15f3f3f5fbed598ae0664e644f84fbf4f2face704a3640dd8a195446f30599f7e4285a38faecf83b4ea61910b0e6c552593b493942490902c4a94d73eab18562a85ebfdfc3ddc51f406d43b89333fc4e5ec25059be7b9f15324160c33579ece90d823b1159bcbec4792dce6bdf31959486e0a168e781631a5b27080b82946a420467d180e30baefe346a0fd6bd06bef91aba66216c3f0dcc709d5b235b62ee98be9370ca3edac7c3b6bb56436dd3b38baa8bf7acc79488467191e40f23f38c72350b9dd829f315c9031be4d653c39fa3b82e69c48f118979849f2f6014087c60e520437076761c9f8c588b791aab9dba197af2b6de5dca8afc1df4fb63400f1b97cc67dcef5b5cb751b892e7d53a302cf8036062b98eebfe6cf4b640ef23d0101ec80e40bb69981fad2f3abac95956a217c12e6ae3d02e2088f3b164d7ab45956b78d40b88d550f46be698db0a190de0a1cc3adec79a55b04eecb4d14509e96b1c1eddd13eadc3a67693f1ed82ebc7505e15d35e1a56a55832a74e421441ec0870199ee946e8e534a4e90bbe0a5d7f9961a8a340aadad63fae6d6d51134c8af",
				},
			},
		},
		{
			configPath: filepath.Join("testdata", "ai_less.yaml"),
			cases: []testCase{
				{
					name:        "Modules Encryption Validation",
					input:       "1",
					expectedHex: "3562b3107539bae5e9ece48f9a0e9c7771b9c37fa1f2ba9de74bc706391851de36225afd0994c2608747879003e285e3e7820209c52541156b870c75489ef2c27799a48a4e71570d054387bcce8a84fc169ef2f1abdada63a085bc0e8d99c6603553579e52c4042f5226d38371c49dd6c21c0e17fcd6fe5fdc31276b553d3e2dc209842d79b1ddbbbf4332c420b8c637c62d73412906361824a3a8d53b3fb53e79b55b440a9b1b98fab604aa759fa28f70acdf02c80a6607cee0614a5d77cd43ad2dfa95f33b02581c4af07e7c76b4addf313a2055fb1a486ad9b1f752c765e71432e7ebb23c33c032776ab35cae34d777f22eeb49ca02d039b7f49b6c9f9a1a6c1a45838ad5ef4530e1c73a15df453444ecbca50f7dc2b80d291f43b3c626646bb62be8eb254540fdf549d1ce49198d283ee9e2143a34aecc4c24da3e6f7e84",
				},
				{
					name:        "STS Encryption Validation",
					input:       "2",
					expectedHex: "a223c8b37136928ac1fcb2b9873a096140ff42526492d690f46a336731a0818ac84ac8165525a0ff9d551fa4381ae0ac580358e4a617449d192ce01f3ebb4ad291c2741c11679850cc512e548c5d9abeb15970685d86c4d71946823b36fdf3682a86c12a496be1134622c9c53bc01ca0025b1f905e8ac7a2d26e06b6f2ec837595b3997c43ab416fc7f36b8e09487ea8f4a0aef4fbaddc166eff09d0a603b4f516123e55112dd6d5ff190fd5fc381c1f68942de1126511277adcc1a3253674ca9ec87eef8e67f1ce7a438b97010ac6b2016e486faaca75c829f8ff4369efd088d5281d4d1724f9da334a6458ddee95197bab3eb90f3eb80cbe843cc3a35d05c37d55d8be42b84a725233b85760e40c1b4273b06cdf5ef996caeccef605a80cbd1406604b16d3b05dcf9d408a054ac7d98e475f185fd0a58fd1d084a0d38ba0b925f22681d98dfb91742cf26a45bc46297d337dbbe26afe28142b753a3b8c9da332cb2c6abe3f87e4123ca7c4733e3b2c8a9dcadb6da776668543fdb42b54e15712be4fc3fea81c9ea28349fad0698514ce7e9bee99df7b31979800ff7e9a81328738bb8bbfcf71aa944f6d7c98250c8a4324bf419127260423e7b2a53057ac9d9ce91a866f82eabacbc270d32558b5e7ae741c51cd2a7c5c00794848f10ccb5d0c4f567621e6b886841d57539416908757a3e4cdf5350610a19481afbe6c4bb923c432cf2a223ee2bd31674f5e02d8a59c733a2205de52441a0ea5be467e08c0115b54f9c783576beff29688ad2d37c068b4d863820e137d0aa9179f6fcae11f0be017211d8ea8456dfe1ada803fd22aceff240b10553e7db0d75128b9ce499919ff74b63118849a6ebf2dcf3d65babf5feb94256d6ee9def987e711c8fd5af234844c9203f5979bf1c6b813a0282db26b2b2e12d96c8b62cfc8014984f711d217247af5196e6b6e5a6d2dcdf1679548585cd458c7a5fe1a45ebe0626f9d4cefe23a593fe05ffd28eac7d73395a2867ced8b594369f1cbdf8d6ac48167f61e750a6df4ac5dfc41e22aa8bd8d83bcd16912c08220d9c83e4acdbdca4be5ed1b5ff4ee5d3f90d5518461d49fde131938193d54534d28c206b44a174407f71ce3c1329949c7b473d5768c5b06edc732d1aa742d847326a93c687bde2b64d891164842e1aa39adcd51e1e7b35baedb507c43e2294fc8fffe288db295aba2253485afd9f794b9ca5610bedf9678cb8015027419a0ce7a7218c5878150c86c4ec6a6d712da03a98ae57b293d2ea87dd65908131c6f8a2df0678358a10946561556f16bff1ecd9c842e908109a381f4f9d34742df0b4db5b1a5038e3a3a8d687b7596607fcc0f39a5b7a23db01e660ceaeffccbc3987258e481eea7556bb411f2810dad7220d5bf14f4e9512ab0d3a24c8c47c40127f50846749dc4c49d048df36179a1c2b81420bc2fd1d0bec970cb75e65d260f53209321559916a2544dcb408aee8d31476ca7169c44b86e13c2941b8fbb45ff7b8741a7481a231b9465e8b84ecf04c3525d8ce4676041e62437f230a3f72fd681725fa05e5574fd2b26f82c62a212639ac05958de49e141a70a2497a089abfd146d9929a5e115b99952b8f9c2d4b78db87be8c541538b0aa770d8bf2e5eedb0aed74dd0da234a3f37433693e8bcb9b587d12cffaa32de7eeaffd3feaa0575408e6ea8bf8f2e465aefb19c633016ac03a1bfca1595486351ebd1cc2d8ede3c939b70f29a75f1b706f25f9ec10811316ee194513837da9824439e47afcbca0bdc914062d1f2e7337a8935e23e69e7201c7286de3b60fb965fcdba91d5b51a804c5abf05f881d7bc0e060d54b45f55220ce1ac9c0f15cd728a5d8627e1304920ad85feb7c52760384a9f42daf9df3724d4d2fee54be055b6b4e4eb508c8c1078a40902a9c6bc435ca1544bfd7fd5085de9fa4bcb22749983505b530e040e71b0f527de4fd9e7a6ca4003dfb8ade013d730410e08b8eeda76cc53ee890c296547",
				},
			},
		},
	}

	// 4. Run the tests.
	for _, tt := range tests {
		cfg, err := config.LoadConfig(tt.configPath)
		if err != nil {
			t.Fatalf("Failed to load config file '%s': %v", tt.configPath, err)
		}

		// 2. Generate INI plaintext from the loaded config.
		modulesPlaintext := cfg.GenerateModulesPlaintext()
		stsPlaintext := cfg.GenerateSTSPlaintext()

		for _, cc := range tt.cases {
			t.Run(cc.name, func(t *testing.T) {
				var got string
				var err error
				switch cc.input {
				case "1":
					got, err = Encrypt(modulesPlaintext)
				case "2":
					got, err = Encrypt(stsPlaintext)
				}

				if err != nil {
					t.Fatalf("Encrypt() returned an unexpected error: %v", err)
				}

				expectedOutput := "#x" + cc.expectedHex
				got = strings.ToLower(got) // Ensure consistent casing

				if got != expectedOutput {
					t.Errorf("Encrypt() mismatch:\ngot:  %v\nwant: %v", got, expectedOutput)
				}
			})
		}
	}
}

type testCase struct {
	name        string
	input       string
	expectedHex string
}

type testConfig struct {
	configPath string
	cases      []testCase
}
