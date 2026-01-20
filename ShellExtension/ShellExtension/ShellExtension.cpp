#include "ShellExtension.h"
#include <shlwapi.h>
#include <strsafe.h>
#include <windows.h>

CShellExt::CShellExt() : m_cRef(1) {}
CShellExt::~CShellExt() {}

ULONG STDMETHODCALLTYPE CShellExt::AddRef() {
    return InterlockedIncrement(&m_cRef);
}

ULONG STDMETHODCALLTYPE CShellExt::Release() {
    ULONG cRef = InterlockedDecrement(&m_cRef);
    if (cRef == 0) delete this;
    return cRef;
}

HRESULT STDMETHODCALLTYPE CShellExt::QueryInterface(REFIID riid, void** ppv) {
    if (IsEqualIID(riid, IID_IUnknown) || IsEqualIID(riid, IID_IShellExtInit)) {
        *ppv = static_cast<IShellExtInit*>(this);
    }
    else if (IsEqualIID(riid, IID_IContextMenu)) {
        *ppv = static_cast<IContextMenu*>(this);
    }
    else {
        *ppv = nullptr;
        return E_NOINTERFACE;
    }
    AddRef();
    return S_OK;
}

HRESULT STDMETHODCALLTYPE CShellExt::Initialize(LPCITEMIDLIST, IDataObject* pDataObj, HKEY) {
    if (!pDataObj) return E_INVALIDARG;

    FORMATETC fmt = { CF_HDROP, nullptr, DVASPECT_CONTENT, -1, TYMED_HGLOBAL };
    STGMEDIUM stg = { 0 };

    if (FAILED(pDataObj->GetData(&fmt, &stg))) return E_FAIL;

    HDROP hDrop = (HDROP)GlobalLock(stg.hGlobal);
    if (hDrop) {
        WCHAR szFile[MAX_PATH];
        if (DragQueryFileW(hDrop, 0, szFile, ARRAYSIZE(szFile))) {
            m_selectedFile = szFile;
        }
        GlobalUnlock(stg.hGlobal);
    }

    ReleaseStgMedium(&stg);
    return S_OK;
}

HRESULT STDMETHODCALLTYPE CShellExt::QueryContextMenu(HMENU hMenu, UINT indexMenu, UINT idCmdFirst,
    UINT idCmdLast, UINT uFlags) {
    if (uFlags & CMF_DEFAULTONLY) return MAKE_HRESULT(SEVERITY_SUCCESS, 0, USHORT(0));

    InsertMenuW(hMenu, indexMenu, MF_BYPOSITION, idCmdFirst + 0, L"Custom Shell Action");
    return MAKE_HRESULT(SEVERITY_SUCCESS, 0, USHORT(1));
}

HRESULT STDMETHODCALLTYPE CShellExt::InvokeCommand(LPCMINVOKECOMMANDINFO pCmdInfo) {
    if (HIWORD(pCmdInfo->lpVerb)) return E_FAIL;

    if (LOWORD(pCmdInfo->lpVerb) == 0) {
        MessageBoxW(nullptr, m_selectedFile.c_str(), L"Selected File", MB_OK);
        return S_OK;
    }

    return E_FAIL;
}

HRESULT STDMETHODCALLTYPE CShellExt::GetCommandString(UINT_PTR, UINT, UINT*, LPSTR, UINT) {
    return E_NOTIMPL;
}
