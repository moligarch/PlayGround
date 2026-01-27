#define _WIN32_DCOM
#include <windows.h>
#include <wbemidl.h>
#include <comdef.h>
#include <iostream>
#include <fstream>
#include <filesystem>
#include <vector>
#include <sstream>
#pragma comment(lib, "wbemuuid.lib")

namespace fs = std::filesystem;

void AppendLog(const fs::path& path, const std::string& msg) {
    try {
        if (!fs::exists(path.parent_path())) {
            fs::create_directories(path.parent_path());
            std::cout << "Created log directory: " << path.parent_path().string() << std::endl;
        }
        std::ofstream f(path, std::ios::app);
        if (f.is_open()) {
            f << msg << std::endl;
            f.close();
            std::cout << "Logged: " << msg << std::endl;
        }
        else {
            std::cerr << "Failed to open log file: " << path << std::endl;
        }
    }
    catch (const std::exception& ex) {
        std::cerr << "Error in AppendLog: " << ex.what() << std::endl;
    }
    catch (...) {
        std::cerr << "Unknown error in AppendLog for file: " << path << std::endl;
    }
}

BOOL IsElevated() {
    BOOL fRet = FALSE;
    HANDLE hToken = nullptr;
    if (OpenProcessToken(GetCurrentProcess(), TOKEN_QUERY, &hToken)) {
        TOKEN_ELEVATION Elevation;
        DWORD cbSize = sizeof(TOKEN_ELEVATION);
        if (GetTokenInformation(hToken, TokenElevation, &Elevation, sizeof(Elevation), &cbSize)) {
            fRet = Elevation.TokenIsElevated;
        }
        CloseHandle(hToken);
    }
    return fRet;
}

#define LOG_HRESULT(hres, msg) \
    if (FAILED(hres)) { \
        std::ostringstream oss; \
        oss << msg << " HRESULT: 0x" << std::hex << hres; \
        AppendLog(LogPath, oss.str()); \
        std::cerr << oss.str() << std::endl; \
    }

int wmain(int argc, wchar_t** argv) {
    if (!IsElevated()) {
        std::cerr << "This program must be run as an administrator!" << std::endl;
        return 1;
    }

    auto DirectoryPath = fs::path{L"C:\\Program Files\\EDR"};
    std::error_code ec{};
    auto cwd = fs::current_path();
    auto LogPath = cwd / "DefenderExclusions.log";

    // Test logging immediately
    AppendLog(LogPath, "Program started at: " + cwd.string());

    bool folderCreated = false;
    bool exclusionAdded = false;

    try {
        if (!fs::exists(DirectoryPath)) {
            fs::create_directories(DirectoryPath);
            std::cout << "Created directory: " << DirectoryPath << std::endl;
            AppendLog(LogPath, "Created directory: " + DirectoryPath.string());
            folderCreated = true;
        }
        else {
            std::cout << "Directory already exists: " << DirectoryPath << std::endl;
            AppendLog(LogPath, "Directory already exists: " + DirectoryPath.string());
            folderCreated = true;
        }
    }
    catch (const std::exception& ex) {
        std::cerr << "Error creating folder: " << ex.what() << std::endl;
        AppendLog(LogPath, std::string("Error creating folder: ") + ex.what());
    }

    HRESULT hres = CoInitializeEx(nullptr, COINIT_MULTITHREADED);
    LOG_HRESULT(hres, "CoInitializeEx failed");
    if (FAILED(hres)) {
        AppendLog(LogPath, "Exiting due to CoInitializeEx failure");
        return 1;
    }

    hres = CoInitializeSecurity(
        nullptr, -1, nullptr, nullptr,
        RPC_C_AUTHN_LEVEL_DEFAULT,
        RPC_C_IMP_LEVEL_IMPERSONATE,
        nullptr, EOAC_NONE, nullptr
    );
    LOG_HRESULT(hres, "CoInitializeSecurity failed");
    if (FAILED(hres)) {
        AppendLog(LogPath, "Exiting due to CoInitializeSecurity failure");
        CoUninitialize();
        return 1;
    }

    IWbemLocator* pLoc = nullptr;
    hres = CoCreateInstance(CLSID_WbemLocator, 0, CLSCTX_INPROC_SERVER, IID_IWbemLocator, (LPVOID*)&pLoc);
    LOG_HRESULT(hres, "CoCreateInstance for IWbemLocator failed");
    if (FAILED(hres) || pLoc == nullptr) {
        AppendLog(LogPath, "Exiting due to CoCreateInstance failure");
        CoUninitialize();
        return 1;
    }

    IWbemServices* pSvc = nullptr;
    hres = pLoc->ConnectServer(_bstr_t(L"ROOT\\Microsoft\\Windows\\Defender"), nullptr, nullptr, nullptr, 0, nullptr, nullptr, &pSvc);
    LOG_HRESULT(hres, "ConnectServer to ROOT\\Microsoft\\Windows\\Defender failed (are you admin?)");
    if (FAILED(hres) || pSvc == nullptr) {
        AppendLog(LogPath, "Exiting due to ConnectServer failure");
        pLoc->Release();
        CoUninitialize();
        return 1;
    }

    hres = CoSetProxyBlanket(pSvc, RPC_C_AUTHN_WINNT, RPC_C_AUTHZ_NONE, nullptr, RPC_C_AUTHN_LEVEL_CALL, RPC_C_IMP_LEVEL_IMPERSONATE, nullptr, EOAC_NONE);
    LOG_HRESULT(hres, "CoSetProxyBlanket failed");
    if (FAILED(hres)) {
        AppendLog(LogPath, "Exiting due to CoSetProxyBlanket failure");
        pSvc->Release();
        pLoc->Release();
        CoUninitialize();
        return 1;
    }

    IWbemClassObject* pClass = nullptr;
    hres = pSvc->GetObject(_bstr_t(L"MSFT_MpPreference"), 0, nullptr, &pClass, nullptr);
    LOG_HRESULT(hres, "GetObject MSFT_MpPreference failed");
    if (FAILED(hres) || !pClass) {
        AppendLog(LogPath, "Exiting due to GetObject failure");
        pSvc->Release();
        pLoc->Release();
        CoUninitialize();
        return 1;
    }

    IWbemClassObject* pInParamsDefinition = nullptr;
    hres = pClass->GetMethod(L"Add", 0, &pInParamsDefinition, nullptr);
    LOG_HRESULT(hres, "GetMethod Add failed");
    if (FAILED(hres) || !pInParamsDefinition) {
        AppendLog(LogPath, "Exiting due to GetMethod failure");
        pClass->Release();
        pSvc->Release();
        pLoc->Release();
        CoUninitialize();
        return 1;
    }

    IWbemClassObject* pInParams = nullptr;
    hres = pInParamsDefinition->SpawnInstance(0, &pInParams);
    LOG_HRESULT(hres, "SpawnInstance for in params failed");
    if (FAILED(hres) || !pInParams) {
        AppendLog(LogPath, "Exiting due to SpawnInstance failure");
        pInParamsDefinition->Release();
        pClass->Release();
        pSvc->Release();
        pLoc->Release();
        CoUninitialize();
        return 1;
    }

    SAFEARRAYBOUND sab;
    sab.lLbound = 0;
    sab.cElements = 1;
    SAFEARRAY* psa = SafeArrayCreate(VT_BSTR, 1, &sab);
    if (!psa) {
        AppendLog(LogPath, "SafeArrayCreate failed");
        std::cerr << "SafeArrayCreate failed" << std::endl;
        pInParams->Release();
        pInParamsDefinition->Release();
        pClass->Release();
        pSvc->Release();
        pLoc->Release();
        CoUninitialize();
        return 1;
    }

    LONG idx = 0;
    BSTR bstrPath = SysAllocString(DirectoryPath.c_str());
    if (!bstrPath) {
        AppendLog(LogPath, "SysAllocString failed");
        std::cerr << "SysAllocString failed" << std::endl;
        SafeArrayDestroy(psa);
        pInParams->Release();
        pInParamsDefinition->Release();
        pClass->Release();
        pSvc->Release();
        pLoc->Release();
        CoUninitialize();
        return 1;
    }

    hres = SafeArrayPutElement(psa, &idx, bstrPath);
    SysFreeString(bstrPath);
    LOG_HRESULT(hres, "SafeArrayPutElement failed");
    if (FAILED(hres)) {
        AppendLog(LogPath, "Exiting due to SafeArrayPutElement failure");
        SafeArrayDestroy(psa);
        pInParams->Release();
        pInParamsDefinition->Release();
        pClass->Release();
        pSvc->Release();
        pLoc->Release();
        CoUninitialize();
        return 1;
    }

    VARIANT var;
    VariantInit(&var);
    var.vt = VT_ARRAY | VT_BSTR;
    var.parray = psa;

    hres = pInParams->Put(L"ExclusionPath", 0, &var, 0);
    LOG_HRESULT(hres, "Put ExclusionPath failed");
    if (FAILED(hres)) {
        SafeArrayDestroy(psa);
        VariantClear(&var);
        pInParams->Release();
        pInParamsDefinition->Release();
        pClass->Release();
        pSvc->Release();
        pLoc->Release();
        CoUninitialize();
        return 1;
    }

    VARIANT vForce;
    VariantInit(&vForce);
    vForce.vt = VT_BOOL;
    vForce.boolVal = VARIANT_TRUE;
    hres = pInParams->Put(L"Force", 0, &vForce, 0);
    LOG_HRESULT(hres, "Put Force failed");
    VariantClear(&vForce);
    if (FAILED(hres)) {
        SafeArrayDestroy(psa);
        VariantClear(&var);
        pInParams->Release();
        pInParamsDefinition->Release();
        pClass->Release();
        pSvc->Release();
        pLoc->Release();
        CoUninitialize();
        return 1;
    }

    IWbemClassObject* pOutParams = nullptr;
    hres = pSvc->ExecMethod(_bstr_t(L"MSFT_MpPreference"), _bstr_t(L"Add"), 0, nullptr, pInParams, &pOutParams, nullptr);
    LOG_HRESULT(hres, "ExecMethod Add failed - maybe not admin or Defender unavailable");
    if (SUCCEEDED(hres) && pOutParams) {
        VARIANT ret;
        VariantInit(&ret);
        hres = pOutParams->Get(L"ReturnValue", 0, &ret, nullptr, nullptr);
        LOG_HRESULT(hres, "Get ReturnValue failed");
        if (SUCCEEDED(hres) && (ret.vt == VT_I4)) {
            long rv = ret.lVal;
            std::ostringstream oss;
            oss << "ExecMethod returned: " << rv;
            AppendLog(LogPath, oss.str());
            std::cout << oss.str() << std::endl;
            if (rv == 0) exclusionAdded = true;
        }
        else {
            AppendLog(LogPath, "Could not read ReturnValue");
            std::cerr << "Could not read ReturnValue" << std::endl;
        }
        VariantClear(&ret);
        pOutParams->Release();
    }

    SafeArrayDestroy(psa);
    VariantClear(&var);
    pInParams->Release();
    pInParamsDefinition->Release();
    pClass->Release();
    pSvc->Release();
    pLoc->Release();
    CoUninitialize();

    std::string result = std::string("FolderCreated=") + (folderCreated ? "True" : "False") +
        ", ExclusionAdded=" + (exclusionAdded ? "True" : "False");
    AppendLog(LogPath, result);
    std::cout << result << std::endl;
    std::cout << "Done. Log: " << LogPath << std::endl;
    return 0;
}