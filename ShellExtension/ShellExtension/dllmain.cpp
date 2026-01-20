#include "ShellExtension.h"
#include <windows.h>
#include <objbase.h>
#include <olectl.h>

HINSTANCE g_hInst;
long g_cDllRef = 0;

class ClassFactory : public IClassFactory {
    long m_cRef;
public:
    ClassFactory() : m_cRef(1) {}
    ~ClassFactory() {}

    IFACEMETHODIMP QueryInterface(REFIID riid, void** ppv) {
        if (IsEqualIID(riid, IID_IUnknown) || IsEqualIID(riid, IID_IClassFactory)) {
            *ppv = static_cast<IClassFactory*>(this);
            AddRef();
            return S_OK;
        }
        *ppv = nullptr;
        return E_NOINTERFACE;
    }

    IFACEMETHODIMP_(ULONG) AddRef() { return InterlockedIncrement(&m_cRef); }
    IFACEMETHODIMP_(ULONG) Release() {
        ULONG cRef = InterlockedDecrement(&m_cRef);
        if (cRef == 0) delete this;
        return cRef;
    }

    IFACEMETHODIMP CreateInstance(IUnknown*, REFIID riid, void** ppv) {
        if (!ppv) return E_POINTER;
        CShellExt* pExt = new CShellExt();
        HRESULT hr = pExt->QueryInterface(riid, ppv);
        pExt->Release();
        return hr;
    }

    IFACEMETHODIMP LockServer(BOOL) { return S_OK; }
};


STDAPI DllGetClassObject(REFCLSID rclsid, REFIID riid, void** ppv) {
    static const CLSID CLSID_ShellExt = { 0x01234567, 0x89AB, 0xCDEF, {0x01,0x23,0x45,0x67,0x89,0xAB,0xCD,0xEF} };
    if (!IsEqualCLSID(rclsid, CLSID_ShellExt)) return CLASS_E_CLASSNOTAVAILABLE;
    ClassFactory* pFactory = new ClassFactory();
    return pFactory->QueryInterface(riid, ppv);
}

STDAPI DllCanUnloadNow() {
    return g_cDllRef == 0 ? S_OK : S_FALSE;
}

STDAPI DllRegisterServer() {
    wchar_t szModule[MAX_PATH];
    if (!GetModuleFileNameW(g_hInst, szModule, ARRAYSIZE(szModule))) {
        return HRESULT_FROM_WIN32(GetLastError());
    }

    const wchar_t* clsid = L"{01234567-89AB-CDEF-0123-456789ABCDEF}";
    HKEY hKey = nullptr;

    // Register under CLSID
    std::wstring keyPath = L"Software\\Classes\\CLSID\\";
    keyPath += clsid;

    if (RegCreateKeyExW(HKEY_LOCAL_MACHINE, keyPath.c_str(), 0, nullptr, 0, KEY_WRITE, nullptr, &hKey, nullptr) != ERROR_SUCCESS) {
        return SELFREG_E_CLASS;
    }

    RegSetValueW(hKey, nullptr, REG_SZ, L"My Shell Extension", 0);

    // InprocServer32
    HKEY hInProc = nullptr;
    if (RegCreateKeyExW(hKey, L"InprocServer32", 0, nullptr, 0, KEY_WRITE, nullptr, &hInProc, nullptr) == ERROR_SUCCESS) {
        RegSetValueExW(hInProc, nullptr, 0, REG_SZ, (BYTE*)szModule, ((DWORD)wcslen(szModule) + 1) * sizeof(wchar_t));
        RegSetValueExW(hInProc, L"ThreadingModel", 0, REG_SZ, (BYTE*)L"Apartment", sizeof(L"Apartment"));
        RegCloseKey(hInProc);
    }
    RegCloseKey(hKey);

    // Register as context menu handler for txt files
    std::wstring handlerKey = L"Software\\Classes\\*\\shellex\\ContextMenuHandlers\\MyShellExt";
    if (RegCreateKeyExW(HKEY_LOCAL_MACHINE, handlerKey.c_str(), 0, nullptr, 0, KEY_WRITE, nullptr, &hKey, nullptr) == ERROR_SUCCESS) {
        RegSetValueExW(hKey, nullptr, 0, REG_SZ, (BYTE*)clsid, ((DWORD)wcslen(clsid) + 1) * sizeof(wchar_t));
        RegCloseKey(hKey);
    }

    SHChangeNotify(SHCNE_ASSOCCHANGED, SHCNF_IDLIST, nullptr, nullptr);
    MessageBoxW(NULL, L"Register the DLL manually or extend this function", L"Stub", MB_OK);
    return S_OK;
}

STDAPI DllUnregisterServer() {
    RegDeleteTreeW(HKEY_LOCAL_MACHINE, L"Software\\Classes\\*\\shellex\\ContextMenuHandlers\\MyShellExt");
    RegDeleteTreeW(HKEY_LOCAL_MACHINE, L"Software\\Classes\\CLSID\\{01234567-89AB-CDEF-0123-456789ABCDEF}");
    SHChangeNotify(SHCNE_ASSOCCHANGED, SHCNF_IDLIST, nullptr, nullptr);
    MessageBoxW(NULL, L"Unregister manually", L"Stub", MB_OK);
    return S_OK;
}

BOOL APIENTRY DllMain(HMODULE hModule, DWORD reason, LPVOID) {
    if (reason == DLL_PROCESS_ATTACH) g_hInst = hModule;
    return TRUE;
}
