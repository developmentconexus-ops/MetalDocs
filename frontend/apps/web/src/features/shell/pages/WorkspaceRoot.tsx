import { useCallback, useEffect, useMemo, useRef } from "react";
import { Outlet, useMatches, useNavigate, useOutletContext } from "react-router-dom";
import * as iamApi from '../../iam/api/iam';
import * as notificationsApiClient from '../../notifications/api/notifications';
import { listTaxonomyAreas, listTaxonomyProfiles } from '../../taxonomy/api/catalog';
import { markUx, startApiTrace, stopApiTrace } from "../../../lib/observability/apiTrace";
import { AuthShell } from "../../../components/AuthShell";
import { PasswordChangePanel } from "../../../components/PasswordChangePanel";
import type { WorkspaceView } from "../../../components/DocumentWorkspaceShell";
import type { CurrentUser, ManagedUserItem } from "../../../lib/types";
import { useDocumentsStore } from "../../documents/state/documents.store";
import { useNotificationsStore } from "../../notifications/state/notifications.store";
import { useRegistryStore } from "../../registry/state/registry.store";
import { useUiStore } from "../../../store/ui.store";
import { useAuthSession } from "../../auth/useAuthSession";
import { useNotifications } from "../../notifications/useNotifications";
import { useRegistryExplorer } from "../../registry/useRegistryExplorer";
import { asMessage } from "../../shared/errors";
import { WorkspaceShell } from "../WorkspaceShell";

type WorkspaceRouteHandle = {
  workspaceView?: WorkspaceView;
  requiresAdmin?: boolean;
  editMode?: boolean;
};

const workspacePathByView: Record<WorkspaceView, string> = {
  operations: "/",
  approvals: "/approvals",
  audit: "/audit",
  library: "/documents",
  "my-docs": "/documents/mine",
  recent: "/documents/recent",
  create: "/documents-v2/new",
  "content-builder": "/content-builder",
  registry: "/registry",
  notifications: "/notifications",
  admin: "/admin",
  "taxonomy-admin": "/admin/taxonomy",
  "templates-v2": "/templates-v2",
  "documents-v2": "/documents-v2/new",
  "registry-v2": "/registry-v2",
  "iam-memberships": "/admin/memberships",
  "approval-routes": "/approval-routes",
};

export type WorkspaceRouteContext = {
  loadState: ReturnType<typeof useDocumentsStore.getState>["loadState"];
  documents: ReturnType<typeof useDocumentsStore.getState>["documents"];
  visibleDocuments: ReturnType<typeof useDocumentsStore.getState>["documents"];
  selectedDocument: ReturnType<typeof useDocumentsStore.getState>["selectedDocument"];
  documentForm: ReturnType<typeof useDocumentsStore.getState>["documentForm"];
  managedUsers: ManagedUserItem[];
  searchQuery: string;
  currentUserId: string;
  isAdmin: boolean;
  documentProfiles: ReturnType<typeof useRegistryStore.getState>["documentProfiles"];
  processAreas: ReturnType<typeof useRegistryStore.getState>["processAreas"];
  documentDepartments: ReturnType<typeof useRegistryStore.getState>["documentDepartments"];
  subjects: ReturnType<typeof useRegistryStore.getState>["subjects"];
  selectedProfileSchema: ReturnType<typeof useRegistryExplorer>["selectedProfileSchema"];
  selectedProfileSchemas: ReturnType<typeof useRegistryExplorer>["selectedProfileSchemas"];
  selectedProfileGovernance: ReturnType<typeof useRegistryExplorer>["selectedProfileGovernance"];
  notifications: ReturnType<typeof useNotifications>["notifications"];
  formatDate: (value?: string) => string;
  onSearchQueryChange: (value: string) => void;
  onCreateDocument: () => void;
  onRefreshWorkspace: () => void | Promise<void>;
  onOpenDocument: (documentId: string, nextView?: "library" | "content-builder") => Promise<void>;
  onOpenDocumentForHub: (documentId: string) => Promise<void>;
  onBackToCreate: () => void;
  onApplyDocumentProfile: ReturnType<typeof useRegistryExplorer>["applyDocumentProfile"];
  onCreateProcessArea: ReturnType<typeof useRegistryExplorer>["handleCreateProcessArea"];
  onUpdateProcessArea: ReturnType<typeof useRegistryExplorer>["handleUpdateProcessArea"];
  onDeleteProcessArea: ReturnType<typeof useRegistryExplorer>["handleDeleteProcessArea"];
  onCreateSubject: ReturnType<typeof useRegistryExplorer>["handleCreateSubject"];
  onUpdateSubject: ReturnType<typeof useRegistryExplorer>["handleUpdateSubject"];
  onDeleteSubject: ReturnType<typeof useRegistryExplorer>["handleDeleteSubject"];
  onCreateDocumentProfile: ReturnType<typeof useRegistryExplorer>["handleCreateDocumentProfile"];
  onUpdateDocumentProfile: ReturnType<typeof useRegistryExplorer>["handleUpdateDocumentProfile"];
  onDeleteDocumentProfile: ReturnType<typeof useRegistryExplorer>["handleDeleteDocumentProfile"];
  onUpdateDocumentProfileGovernance: ReturnType<typeof useRegistryExplorer>["handleUpdateDocumentProfileGovernance"];
  onUpsertDocumentProfileSchema: ReturnType<typeof useRegistryExplorer>["handleUpsertDocumentProfileSchema"];
  onActivateDocumentProfileSchema: ReturnType<typeof useRegistryExplorer>["handleActivateDocumentProfileSchema"];
  onMarkNotificationRead: ReturnType<typeof useNotifications>["handleMarkNotificationRead"];
  navigateToWorkspaceView: (view: WorkspaceView) => void;
};

export function useWorkspaceRouteContext() {
  return useOutletContext<WorkspaceRouteContext>();
}

export function Component() {
  const matches = useMatches();
  const navigate = useNavigate();
  const routeHandle = [...matches].reverse().find((match) => (match.handle as WorkspaceRouteHandle | undefined)?.workspaceView)?.handle as WorkspaceRouteHandle | undefined;
  const activeView = routeHandle?.workspaceView ?? "operations";
  const requiresAdmin = routeHandle?.requiresAdmin ?? false;
  const editMode = routeHandle?.editMode ?? false;

  const {
    message,
    error,
    searchQuery,
    managedUsers,
    setMessage,
    setError,
    setActiveView,
    setSearchQuery,
    setManagedUsers,
  } = useUiStore();

  const {
    loadState,
    documentForm,
    documents,
    selectedDocument,
    setDocumentForm,
    setCollaborationPresence,
    setDocumentEditLock,
    setDocuments,
    setLoadState,
    setSelectedDocument,
    setVersions,
    setApprovals,
    setAttachments,
    setAuditEvents,
    setPolicies,
    setVersionDiff,
    setPolicyResourceId,
  } = useDocumentsStore();
  const { setDocumentProfiles, setProcessAreas, setDocumentDepartments, setSubjects } = useRegistryStore();
  const { setNotifications } = useNotificationsStore();
  const setErrorStore = setError;

  const streamRefreshInFlightRef = useRef(false);
  const registry = useRegistryExplorer(() => refreshWorkspace(user));
  const notificationsApi = useNotifications();

  const loadWorkspace = useCallback(
    async (currentUser: CurrentUser) => {
      setLoadState("loading");
      try {
        const empty = { items: [] };
        const safe = <T,>(p: Promise<T>, fallback: T) => p.catch(() => fallback);
        const currentUserRoleValues = Array.isArray(currentUser.roles) ? currentUser.roles as readonly string[] : [];
        const [profilesResponse, processAreasResponse, departmentsResponse, subjectsResponse, docsResponse, usersResponse, notificationsResponse] = await Promise.all([
          safe(listTaxonomyProfiles(), empty as never),
          safe(listTaxonomyAreas(), empty as never),
          Promise.resolve({ items: [] }),
          Promise.resolve({ items: [] }),
          Promise.resolve({ items: [] }),
          currentUserRoleValues.some((role) => role === "admin" || role === "system_admin")
            ? safe(iamApi.listUsers(), empty as never)
            : Promise.resolve({ items: [] as ManagedUserItem[] }),
          safe(notificationsApiClient.listNotifications(new URLSearchParams({ limit: "10" })), empty as never),
        ]);
        const profiles = Array.isArray(profilesResponse.items) ? profilesResponse.items : [];
        const areas = Array.isArray(processAreasResponse.items) ? processAreasResponse.items : [];
        const departments = Array.isArray(departmentsResponse.items) ? departmentsResponse.items : [];
        const nextSubjects = Array.isArray(subjectsResponse.items) ? subjectsResponse.items : [];
        const docs = Array.isArray(docsResponse.items) ? docsResponse.items : [];
        const users = Array.isArray(usersResponse.items) ? usersResponse.items : [];
        setDocumentProfiles(profiles);
        setProcessAreas(areas);
        setDocumentDepartments(departments);
        setSubjects(nextSubjects);
        setDocuments(docs);
        setManagedUsers(users);
        setNotifications(Array.isArray(notificationsResponse.items) ? notificationsResponse.items : []);
        if (profiles.length > 0) {
          const nextProfileCode = profiles.find((item) => item.code === documentForm.documentProfile)?.code ?? profiles[0]?.code ?? "";
          if (nextProfileCode) {
            try {
              await registry.applyDocumentProfile(nextProfileCode, documentForm.processArea);
              await registry.prefetchProfile(nextProfileCode);
            } catch {
              // profile bundle unavailable (v2-only profile) - workspace still loads
            }
          }
        }
        setLoadState("ready");
      } catch (err) {
        setErrorStore(asMessage(err));
        setLoadState("error");
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [documentForm.documentProfile, documentForm.processArea, setDocuments, setErrorStore, setLoadState, setNotifications, setDocumentProfiles, setProcessAreas, setDocumentDepartments, setSubjects, setManagedUsers],
  );

  const refreshWorkspace = useCallback(async (currentUser: typeof user) => {
    if (!currentUser) return;
    await loadWorkspace(currentUser);
  }, [loadWorkspace]);

  const authSession = useAuthSession({ onAuthenticated: loadWorkspace });
  const { authState, user, loginForm, passwordForm, setLoginForm, setPasswordForm, bootstrap, handleLogin, handleLogout, handleChangePassword } = authSession;

  const handleRefreshWorkspace = useCallback(() => refreshWorkspace(user), [refreshWorkspace, user]);

  const refreshOperationalSignals = useCallback(async () => {
    if (streamRefreshInFlightRef.current) return;
    streamRefreshInFlightRef.current = true;
    try {
      const [docsResponse, notificationsResponse] = await Promise.all([
        Promise.resolve({ items: [] }),
        notificationsApiClient.listNotifications(new URLSearchParams({ limit: "10" })),
      ]);
      setDocuments(Array.isArray(docsResponse.items) ? docsResponse.items : []);
      setNotifications(Array.isArray(notificationsResponse.items) ? notificationsResponse.items : []);
    } finally {
      streamRefreshInFlightRef.current = false;
    }
  }, [setDocuments, setNotifications]);

  const loadDocumentDetails = useCallback(
    async (documentId: string) => {
      const normalizedDocumentID = documentId.trim();
      if (!normalizedDocumentID) {
        setErrorStore("Selecione um documento valido antes de abrir.");
        return false;
      }
      startApiTrace(`open-document:${normalizedDocumentID}`);
      markUx(`open-document-start:${normalizedDocumentID}`);
      setSelectedDocument(null);
      setVersions([]);
      setApprovals([]);
      setAttachments([]);
      setCollaborationPresence([]);
      setDocumentEditLock(null);
      setAuditEvents([]);
      setPolicyResourceId(normalizedDocumentID);
      setPolicies([]);
      setVersionDiff(null);
      stopApiTrace();
      return false;
    },
    [setApprovals, setAttachments, setAuditEvents, setCollaborationPresence, setDocumentEditLock, setErrorStore, setPolicies, setPolicyResourceId, setSelectedDocument, setVersionDiff, setVersions],
  );

  const navigateToWorkspaceView = useCallback((view: WorkspaceView) => {
    navigate(workspacePathByView[view]);
  }, [navigate]);

  const openDocument = useCallback(
    async (documentId: string, nextView: "library" | "content-builder" = "library") => {
      const ok = await loadDocumentDetails(documentId);
      if (ok) navigateToWorkspaceView(nextView);
    },
    [loadDocumentDetails, navigateToWorkspaceView],
  );

  const openDocumentForHub = useCallback(
    async (documentId: string) => {
      const { documents: currentDocs } = useDocumentsStore.getState();
      const existing = currentDocs.find((d) => d.documentId === documentId);
      if (existing) {
        setSelectedDocument(existing);
        return;
      }
      await loadDocumentDetails(documentId);
    },
    [loadDocumentDetails, setSelectedDocument],
  );

  const handleBackToCreate = useCallback(() => {
    if (selectedDocument) {
      setDocumentForm((current) => ({
        ...current,
        title: selectedDocument.title,
        documentType: selectedDocument.documentType,
        documentProfile: selectedDocument.documentProfile,
        processArea: selectedDocument.processArea ?? "",
        subject: selectedDocument.subject ?? "",
        ownerId: selectedDocument.ownerId,
        businessUnit: selectedDocument.businessUnit,
        department: selectedDocument.department,
        classification: selectedDocument.classification,
        tags: selectedDocument.tags.join(", "),
        effectiveAt: selectedDocument.effectiveAt ?? "",
        expiryAt: selectedDocument.expiryAt ?? "",
      }));
    }
    navigateToWorkspaceView("create");
  }, [navigateToWorkspaceView, selectedDocument, setDocumentForm]);

  const {
    applyDocumentProfile,
    handleCreateProcessArea,
    handleUpdateProcessArea,
    handleDeleteProcessArea,
    handleCreateSubject,
    handleUpdateSubject,
    handleDeleteSubject,
    handleCreateDocumentProfile,
    handleUpdateDocumentProfile,
    handleDeleteDocumentProfile,
    handleUpdateDocumentProfileGovernance,
    handleUpsertDocumentProfileSchema,
    handleActivateDocumentProfileSchema,
    documentProfiles,
    processAreas,
    documentDepartments,
    subjects,
    selectedProfileSchema,
    selectedProfileSchemas,
    selectedProfileGovernance,
  } = registry;
  const { notifications, handleMarkNotificationRead, subscribeOperations } = notificationsApi;
  const currentUserRoles = Array.isArray(user?.roles) ? user.roles as readonly string[] : [];
  const isAdmin = currentUserRoles.some((role) => role === "admin" || role === "system_admin");
  const userRoleLabel = roleLabelFromRoles(currentUserRoles);
  const visibleDocuments = activeView === "my-docs"
    ? documents.filter((item) => item.ownerId === user?.userId)
    : activeView === "recent"
      ? [...documents].sort((left, right) => new Date(right.createdAt).getTime() - new Date(left.createdAt).getTime())
      : documents;

  useEffect(() => {
    setActiveView(activeView);
  }, [activeView, setActiveView]);

  useEffect(() => {
    if (requiresAdmin && !isAdmin) {
      navigate("/", { replace: true });
    }
  }, [isAdmin, navigate, requiresAdmin]);

  useEffect(() => {
    void bootstrap();
  }, [bootstrap]);

  useEffect(() => {
    if (authState !== "ready" || !user || user.mustChangePassword) {
      return;
    }
    return subscribeOperations(() => {
      void refreshOperationalSignals();
    });
  }, [authState, refreshOperationalSignals, subscribeOperations, user?.mustChangePassword, user?.userId]);

  useEffect(() => {
    if (!message) {
      return;
    }
    const timer = window.setTimeout(() => {
      setMessage("");
    }, 2000);
    return () => {
      window.clearTimeout(timer);
    };
  }, [message, setMessage]);

  useEffect(() => {
    if (!error) {
      return;
    }
    const timer = window.setTimeout(() => {
      setError("");
    }, 6000);
    return () => {
      window.clearTimeout(timer);
    };
  }, [error, setError]);

  useEffect(() => {
    if (authState !== "ready" || !selectedDocument?.documentId) {
      return;
    }

    const emitHeartbeat = async () => {
      try {
        setCollaborationPresence([]);
        setDocumentEditLock(null);
      } catch {
        // Collaboration refresh is best-effort and must not block normal workspace usage.
      }
    };

    void emitHeartbeat();
    const timer = window.setInterval(() => {
      void emitHeartbeat();
    }, 30000);
    return () => {
      window.clearInterval(timer);
    };
  }, [authState, selectedDocument?.documentId, setCollaborationPresence, setDocumentEditLock, user?.displayName]);

  const routeContext = useMemo<WorkspaceRouteContext>(() => ({
    loadState,
    documents,
    visibleDocuments,
    selectedDocument,
    documentForm,
    managedUsers,
    searchQuery,
    currentUserId: user?.userId ?? "",
    isAdmin,
    documentProfiles,
    processAreas,
    documentDepartments,
    subjects,
    selectedProfileSchema,
    selectedProfileSchemas,
    selectedProfileGovernance,
    notifications,
    formatDate,
    onSearchQueryChange: setSearchQuery,
    onCreateDocument: () => navigateToWorkspaceView("create"),
    onRefreshWorkspace: handleRefreshWorkspace,
    onOpenDocument: openDocument,
    onOpenDocumentForHub: openDocumentForHub,
    onBackToCreate: handleBackToCreate,
    onApplyDocumentProfile: applyDocumentProfile,
    onCreateProcessArea: handleCreateProcessArea,
    onUpdateProcessArea: handleUpdateProcessArea,
    onDeleteProcessArea: handleDeleteProcessArea,
    onCreateSubject: handleCreateSubject,
    onUpdateSubject: handleUpdateSubject,
    onDeleteSubject: handleDeleteSubject,
    onCreateDocumentProfile: handleCreateDocumentProfile,
    onUpdateDocumentProfile: handleUpdateDocumentProfile,
    onDeleteDocumentProfile: handleDeleteDocumentProfile,
    onUpdateDocumentProfileGovernance: handleUpdateDocumentProfileGovernance,
    onUpsertDocumentProfileSchema: handleUpsertDocumentProfileSchema,
    onActivateDocumentProfileSchema: handleActivateDocumentProfileSchema,
    onMarkNotificationRead: handleMarkNotificationRead,
    navigateToWorkspaceView,
  }), [applyDocumentProfile, documentDepartments, documentForm, documentProfiles, documents, handleActivateDocumentProfileSchema, handleBackToCreate, handleCreateDocumentProfile, handleCreateProcessArea, handleCreateSubject, handleDeleteDocumentProfile, handleDeleteProcessArea, handleDeleteSubject, handleMarkNotificationRead, handleRefreshWorkspace, handleUpdateDocumentProfile, handleUpdateDocumentProfileGovernance, handleUpdateProcessArea, handleUpdateSubject, handleUpsertDocumentProfileSchema, isAdmin, loadState, managedUsers, navigateToWorkspaceView, notifications, openDocument, openDocumentForHub, processAreas, searchQuery, selectedDocument, selectedProfileGovernance, selectedProfileSchema, selectedProfileSchemas, setSearchQuery, subjects, user?.userId, visibleDocuments]);

  if (authState === "loading") {
    return <div className="app-shell"><section className="hero-panel"><strong>Validando sessao...</strong></section></div>;
  }

  if (!user) {
    return <AuthShell identifier={loginForm.identifier} password={loginForm.password} message={message} error={error} onIdentifierChange={(identifier) => setLoginForm({ ...loginForm, identifier })} onPasswordChange={(password) => setLoginForm({ ...loginForm, password })} onSubmit={handleLogin} />;
  }

  return (
    <div className={`app-shell ${!user.mustChangePassword ? "is-workspace" : ""}`}>
      {(error || message) && (
        <div className="toast-container" aria-live="polite" aria-atomic="false">
          {error && (
            <div data-testid="app-toast-error" className="toast toast-error" role="alert">
              <span className="toast-icon" aria-hidden="true">!</span>
              <div className="toast-body">
                <strong>Falha</strong>
                <span>{error}</span>
              </div>
              <button type="button" className="toast-close" aria-label="Fechar alerta" onClick={() => setError("")}>x</button>
            </div>
          )}
          {message && (
            <div data-testid="app-toast" className="toast toast-success" role="status">
              <span className="toast-icon" aria-hidden="true">v</span>
              <div className="toast-body">
                <strong>Pronto</strong>
                <span>{message}</span>
              </div>
              <button type="button" className="toast-close" aria-label="Fechar mensagem" onClick={() => setMessage("")}>x</button>
            </div>
          )}
        </div>
      )}

      {user.mustChangePassword && (
        <PasswordChangePanel newPassword={passwordForm.newPassword} confirmPassword={passwordForm.confirmPassword} onNewPasswordChange={(newPassword) => setPasswordForm({ ...passwordForm, newPassword })} onConfirmPasswordChange={(confirmPassword) => setPasswordForm({ ...passwordForm, confirmPassword })} onSubmit={handleChangePassword} />
      )}

      {!user.mustChangePassword && (
        <WorkspaceShell
          userDisplayName={user.displayName}
          userRoleLabel={userRoleLabel}
          activeView={activeView}
          searchValue={searchQuery}
          notificationsPending={notifications.filter((item) => item.status !== "READ").length}
          documentCount={documents.length}
          reviewCount={documents.filter((item) => item.status === "IN_REVIEW").length}
          registryCount={documentProfiles.length}
          showAdmin={isAdmin}
          documentProfiles={documentProfiles}
          processAreas={processAreas}
          documents={documents}
          onSearchChange={setSearchQuery}
          onNavigate={navigateToWorkspaceView}
          onPrimaryAction={() => navigateToWorkspaceView("create")}
          onRefreshWorkspace={handleRefreshWorkspace}
          isRefreshing={loadState === "loading"}
          flushContent={editMode}
          editMode={editMode}
          onLogout={handleLogout}
        >
          <Outlet context={routeContext} />
        </WorkspaceShell>
      )}
    </div>
  );
}

function formatDate(value?: string): string {
  if (!value) return "-";
  return new Intl.DateTimeFormat("pt-BR", { dateStyle: "short", timeStyle: "short" }).format(new Date(value));
}

function roleLabelFromRoles(roles: readonly string[]): string {
  if (roles.includes("admin") || roles.includes("system_admin")) return "Administrador";
  if (roles.includes("approver") || roles.includes("reviewer")) return "Aprovador";
  if (roles.includes("author")) return "Autor";
  if (roles.includes("editor")) return "Editor";
  return "Visualizador";
}
