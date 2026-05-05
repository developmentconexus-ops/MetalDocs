import { Component, useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { api, markUx, startApiTrace, stopApiTrace } from "./lib.api";
import { AuthShell } from "./components/AuthShell";
import { AdminCenterView } from "./features/iam/AdminCenterView";
import { TaxonomyAdminPage } from "./features/taxonomy";
import { NotificationsPanel } from "./components/NotificationsPanel";
import { OperationsCenter } from "./components/OperationsCenter";
import { PasswordChangePanel } from "./components/PasswordChangePanel";
import { ContentBuilderView } from "./components/content-builder/ContentBuilderView";
import { WorkspacePlaceholder } from "./components/WorkspacePlaceholder";
import type { UserRole } from "./lib.types";
import { useUiStore } from "./store/ui.store";
import { useAuthSession } from "./features/auth/useAuthSession";
import { useRegistryExplorer } from "./features/registry/useRegistryExplorer";
import { useDocumentsStore } from "./store/documents.store";
import { useRegistryStore } from "./store/registry.store";
import { useNotificationsStore } from "./store/notifications.store";
import { listTaxonomyAreas, listTaxonomyProfiles } from "./api/registry";
import { asMessage, statusOf } from "./features/shared/errors";
import type { AccessPolicyItem, CurrentUser, ManagedUserItem } from "./lib.types";
import { useNotifications } from "./features/notifications/useNotifications";
import { DocumentsHubView } from "./features/documents/DocumentsHubView";
import { RegistryExplorerView } from "./features/registry/RegistryExplorerView";
import { WorkspaceShell } from "./features/shell/WorkspaceShell";
import { isPathForView, pathFromView, viewFromPath } from "./routing/workspaceRoutes";
import { TemplatesV2View, type TemplatesV2Route } from "./features/templates/v2/routes";
import { renderDocumentsV2View, routeFromPath as docsRouteFromPath, pathFromRoute as docsPathFromRoute, type DocumentsV2Route } from "./features/documents/v2/routes";
import { RegistryListPage } from "./features/registry";
import { AreaMembershipAdminPage } from "./features/iam/AreaMembershipAdminPage";
import { InboxPage } from "./features/approval/pages/InboxPage";
import { RouteAdminPage } from "./features/approval/pages/RouteAdminPage";
import { Toaster } from "sonner";

type AppErrorBoundaryState = {
  hasError: boolean;
  message: string;
};

class AppErrorBoundary extends Component<{ children: React.ReactNode }, AppErrorBoundaryState> {
  state: AppErrorBoundaryState = {
    hasError: false,
    message: "",
  };

  static getDerivedStateFromError(error: Error): AppErrorBoundaryState {
    return {
      hasError: true,
      message: error.message || "Falha inesperada ao renderizar a interface.",
    };
  }

  componentDidCatch(error: Error): void {
    console.error("MetalDocs UI render error", error);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="app-shell">
          <section className="hero-panel stack">
            <strong>Falha ao montar a interface.</strong>
            <p className="hint">{this.state.message}</p>
            <p className="hint">A API respondeu, mas a interface encontrou um dado inesperado durante o render. Recarregue a pagina apos atualizar o frontend local.</p>
          </section>
        </div>
      );
    }
    return this.props.children;
  }
}

export default function App() {
  return (
    <AppErrorBoundary>
      <AppContent />
      <Toaster position="bottom-right" richColors closeButton />
    </AppErrorBoundary>
  );
}

function AppContent() {
  const {
    message,
    error,
    activeView,
    pendingViewNavigation,
    searchQuery,
    managedUsers,
    setMessage,
    setError,
    setActiveView,
    requestViewNavigation,
    clearPendingViewNavigation,
    setSearchQuery,
    setManagedUsers,
  } = useUiStore();

  const location = useLocation();
  const navigate = useNavigate();
  const navSourceRef = useRef<"url" | "store" | null>(null);

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
  // registry and notificationsApi declared before loadWorkspace so their methods are in scope.
  // The callbacks are lazy closures — by call time all vars are initialized.
  const registry = useRegistryExplorer(() => refreshWorkspace(user));
  const notificationsApi = useNotifications();

  const loadWorkspace = useCallback(
    async (currentUser: CurrentUser) => {
      setLoadState("loading");
      try {
        const empty = { items: [] };
        const safe = <T,>(p: Promise<T>, fallback: T) => p.catch(() => fallback);
        const [profilesResponse, processAreasResponse, departmentsResponse, subjectsResponse, docsResponse, usersResponse, notificationsResponse] = await Promise.all([
          safe(listTaxonomyProfiles(), empty as never),
          safe(listTaxonomyAreas(), empty as never),
          Promise.resolve({ items: [] }),
          Promise.resolve({ items: [] }),
          safe(api.searchDocuments(new URLSearchParams({ limit: "25" })), empty as never),
          (Array.isArray(currentUser.roles) ? currentUser.roles : []).includes("admin")
            ? safe(api.listUsers(), empty as never)
            : Promise.resolve({ items: [] as ManagedUserItem[] }),
          safe(api.listNotifications(new URLSearchParams({ limit: "10" })), empty as never),
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
              // profile bundle unavailable (v2-only profile) — workspace still loads
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

  // Zero-arg wrapper for prop sites that declare onRefreshWorkspace: () => void | Promise<void>
  const handleRefreshWorkspace = useCallback(() => refreshWorkspace(user), [refreshWorkspace, user]);

  const refreshOperationalSignals = useCallback(async () => {
    if (streamRefreshInFlightRef.current) return;
    streamRefreshInFlightRef.current = true;
    try {
      const [docsResponse, notificationsResponse] = await Promise.all([
        api.searchDocuments(new URLSearchParams({ limit: "25" })),
        api.listNotifications(new URLSearchParams({ limit: "10" })),
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
      try {
        const [docResponse, versionsResponse, approvalsResponse, attachmentsResponse, auditResponse] = await Promise.all([
          api.getDocument(normalizedDocumentID),
          api.listVersions(normalizedDocumentID),
          api.listApprovals(normalizedDocumentID),
          api.listAttachments(normalizedDocumentID),
          api.listAuditEvents(new URLSearchParams({ resourceId: normalizedDocumentID })),
        ]);
        const orderedVersions = [...versionsResponse.items].sort((a, b) => b.version - a.version);
        setSelectedDocument(docResponse);
        setVersions(orderedVersions);
        setApprovals(approvalsResponse.items);
        setAttachments(attachmentsResponse.items);
        setCollaborationPresence([]);
        setDocumentEditLock(null);
        setAuditEvents(auditResponse.items);
        setPolicyResourceId(normalizedDocumentID);
        const policyResponse = await api.listAccessPolicies("document", normalizedDocumentID).catch((err) => {
          if (statusOf(err) === 403) return { items: [] as AccessPolicyItem[] };
          throw err;
        });
        const nextDiff = orderedVersions.length >= 2
          ? await api.getVersionDiff(normalizedDocumentID, orderedVersions[1].version, orderedVersions[0].version)
          : null;
        setPolicies(policyResponse.items);
        setVersionDiff(nextDiff);
        markUx(`open-document-ready:${normalizedDocumentID}`);
        stopApiTrace();
        return true;
      } catch (err) {
        setErrorStore(asMessage(err));
        stopApiTrace();
        return false;
      }
    },
    [setApprovals, setAttachments, setAuditEvents, setCollaborationPresence, setDocumentEditLock, setErrorStore, setPolicies, setPolicyResourceId, setSelectedDocument, setVersionDiff, setVersions],
  );

  const openDocument = useCallback(
    async (documentId: string, nextView: "library" | "content-builder" = "library") => {
      const ok = await loadDocumentDetails(documentId);
      if (ok) requestViewNavigation(nextView);
    },
    [loadDocumentDetails, requestViewNavigation],
  );

  const openDocumentForHub = useCallback(
    async (documentId: string) => {
      const { documents: currentDocs } = useDocumentsStore.getState();
      const existing = currentDocs.find((d) => d.documentId === documentId);
      if (existing) { setSelectedDocument(existing); return; }
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
    requestViewNavigation("create");
  }, [requestViewNavigation, selectedDocument, setDocumentForm]);
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
  const currentUserRoles = Array.isArray(user?.roles) ? user.roles : [];
  const isAdmin = currentUserRoles.includes("admin") || currentUserRoles.includes("system_admin");
  const userRoleLabel = roleLabelFromRoles(currentUserRoles);
  const visibleDocuments = activeView === "my-docs"
    ? documents.filter((item) => item.ownerId === user?.userId)
    : activeView === "recent"
      ? [...documents].sort((left, right) => new Date(right.createdAt).getTime() - new Date(left.createdAt).getTime())
      : documents;

  const locationView = useMemo(() => viewFromPath(location.pathname), [location.pathname]);
  const [tplRoute, setTplRoute] = useState<TemplatesV2Route>({ kind: 'list' });
  const [docsRoute, setDocsRouteState] = useState<DocumentsV2Route>(() => docsRouteFromPath(location.pathname));
  const setDocsRoute = useCallback((next: DocumentsV2Route) => {
    setDocsRouteState(next);
    const nextPath = docsPathFromRoute(next);
    if (window.location.pathname !== nextPath) {
      window.history.pushState({}, "", nextPath);
    }
  }, []);

  useEffect(() => {
    setDocsRouteState(docsRouteFromPath(location.pathname));
  }, [location.pathname]);

  useEffect(() => {
    const onPopState = () => {
      setDocsRouteState(docsRouteFromPath(window.location.pathname));
    };
    window.addEventListener("popstate", onPopState);
    return () => {
      window.removeEventListener("popstate", onPopState);
    };
  }, []);

  const handleWorkspaceNavigate = useCallback((nextView: Parameters<typeof pathFromView>[0]) => {
    if (nextView === "admin" && !isAdmin) {
      navigate("/", { replace: true });
      return;
    }
    navigate(pathFromView(nextView));
  }, [isAdmin, navigate]);

  const handlePrimaryAction = useCallback(() => {
    navigate(pathFromView("create"));
  }, [navigate]);

  useEffect(() => {
    if (pendingViewNavigation) {
      if (locationView === pendingViewNavigation) {
        clearPendingViewNavigation();
      }
      return;
    }

    if (locationView === "admin" && !isAdmin) {
      if (location.pathname !== "/") {
        navSourceRef.current = "url";
        navigate("/", { replace: true });
      }
      return;
    }

    if (locationView !== activeView) {
      navSourceRef.current = "url";
      setActiveView(locationView);
    }
  }, [activeView, clearPendingViewNavigation, isAdmin, location.pathname, locationView, navigate, pendingViewNavigation, setActiveView]);

  useEffect(() => {
    if (pendingViewNavigation) {
      const targetPath = pathFromView(pendingViewNavigation);
      if (targetPath !== location.pathname) {
        navigate(targetPath);
      }
      return;
    }

    if (navSourceRef.current === "url") {
      navSourceRef.current = null;
      return;
    }

    if (isPathForView(location.pathname, activeView)) {
      return;
    }

    const targetPath = pathFromView(activeView);
    if (targetPath !== location.pathname) {
      navSourceRef.current = "store";
      navigate(targetPath);
    }
  }, [activeView, location.pathname, navigate, pendingViewNavigation]);

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
  }, [message]);

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
  }, [error]);

  useEffect(() => {
    if (authState !== "ready" || !selectedDocument?.documentId) {
      return;
    }

    const emitHeartbeat = async () => {
      try {
        await api.heartbeatDocumentPresence(selectedDocument.documentId, { displayName: user?.displayName ?? "" });
        const [presenceResponse, lockResponse] = await Promise.all([
          api.listDocumentPresence(selectedDocument.documentId),
          api.getDocumentEditLock(selectedDocument.documentId).catch((err) => {
            if (statusOf(err) === 404) {
              return null;
            }
            throw err;
          }),
        ]);
        setCollaborationPresence(presenceResponse.items);
        setDocumentEditLock(lockResponse);
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
  }, [authState, selectedDocument?.documentId, user?.displayName]);

  if (authState === "loading") {
    return <div className="app-shell"><section className="hero-panel"><strong>Validando sessao...</strong></section></div>;
  }

  if (!user) {
    return <AuthShell identifier={loginForm.identifier} password={loginForm.password} message={message} error={error} onIdentifierChange={(identifier) => setLoginForm({ ...loginForm, identifier })} onPasswordChange={(password) => setLoginForm({ ...loginForm, password })} onSubmit={handleLogin} />;
  }

  function renderWorkspaceView() {
    if (activeView === "approvals") {
      return <InboxPage />;
    }

    if (activeView === "approval-routes" && isAdmin) {
      return <RouteAdminPage />;
    }

    if (activeView === "operations" || activeView === "audit") {
      return (
        <OperationsCenter
          loadState={loadState}
          documents={documents}
          notifications={notifications}
          documentProfiles={documentProfiles}
          processAreas={processAreas}
          formatDate={formatDate}
          onRefreshWorkspace={handleRefreshWorkspace}
          onOpenDocument={openDocument}
        />
      );
    }

    if (activeView === "library" || activeView === "my-docs" || activeView === "recent") {
      return (
        <DocumentsHubView
          view={activeView}
          loadState={loadState}
          documentProfiles={documentProfiles}
          processAreas={processAreas}
          documents={visibleDocuments}
          managedUsers={managedUsers}
          selectedDocument={selectedDocument}
          selectedProfileGovernance={selectedProfileGovernance}
          searchQuery={searchQuery}
          currentUserId={user?.userId}
          formatDate={formatDate}
          onSearchQueryChange={setSearchQuery}
          onCreateDocument={handlePrimaryAction}
          onRefreshDocuments={handleRefreshWorkspace}
          onOpenDocument={openDocument}
          onOpenDocumentForHub={openDocumentForHub}
        />
      );
    }

    if (activeView === "content-builder") {
      return (
        <ContentBuilderView
          document={selectedDocument}
          onBack={handleBackToCreate}
          currentUserId={user?.userId ?? ""}
        />
      );
    }

    if (activeView === "registry") {
      return (
        <RegistryExplorerView
          loadState={loadState}
          documentProfiles={documentProfiles}
          processAreas={processAreas}
          subjects={subjects}
          selectedProfileCode={documentForm.documentProfile}
          selectedProfileSchema={selectedProfileSchema}
          selectedProfileSchemas={selectedProfileSchemas}
          selectedProfileGovernance={selectedProfileGovernance}
          showAdmin={isAdmin}
          onRefreshWorkspace={handleRefreshWorkspace}
          onSelectProfile={(profileCode) => applyDocumentProfile(profileCode, documentForm.processArea)}
          onCreateProcessArea={handleCreateProcessArea}
          onUpdateProcessArea={handleUpdateProcessArea}
          onDeleteProcessArea={handleDeleteProcessArea}
          onCreateSubject={handleCreateSubject}
          onUpdateSubject={handleUpdateSubject}
          onDeleteSubject={handleDeleteSubject}
          onCreateDocumentProfile={handleCreateDocumentProfile}
          onUpdateDocumentProfile={handleUpdateDocumentProfile}
          onDeleteDocumentProfile={handleDeleteDocumentProfile}
          onUpdateDocumentProfileGovernance={handleUpdateDocumentProfileGovernance}
          onUpsertDocumentProfileSchema={handleUpsertDocumentProfileSchema}
          onActivateDocumentProfileSchema={handleActivateDocumentProfileSchema}
        />
      );
    }

    if (activeView === "notifications") {
      return (
        <NotificationsPanel
          loadState={loadState}
          notifications={notifications}
          formatDate={formatDate}
          onRefreshWorkspace={handleRefreshWorkspace}
          onMarkRead={handleMarkNotificationRead}
        />
      );
    }

    if (activeView === "admin" && isAdmin) {
      return (
        <AdminCenterView />
      );
    }

    if (activeView === "taxonomy-admin" && isAdmin) {
      return <TaxonomyAdminPage />;
    }

    if (activeView === "templates-v2") {
      return <TemplatesV2View route={tplRoute} onNavigate={setTplRoute} />;
    }
    if (activeView === "documents-v2") {
      return renderDocumentsV2View(docsRoute, setDocsRoute, () => navigate('/registry-v2'));
    }

    if (activeView === "registry-v2") {
      return (
        <RegistryListPage
          onOpenDocumentEditor={(docId) => navigate(docsPathFromRoute({ kind: 'editor', documentID: docId }))}
        />
      );
    }

    if (activeView === "iam-memberships" && isAdmin) {
      return <AreaMembershipAdminPage />;
    }

    return (
      <WorkspacePlaceholder
        kicker="Workspace"
        title="Workspace"
        description="Selecione uma visao operacional na barra lateral para continuar."
        bullets={[
          "Acesse Documentos para explorar o acervo e revisar detalhes.",
          "Use Novo documento para iniciar o fluxo Profile -> Metadata -> Content -> Review.",
          "Abra Tipos documentais para consultar regras e governanca por perfil.",
        ]}
      />
    );
  }

  const workspaceView = renderWorkspaceView();

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
          onNavigate={handleWorkspaceNavigate}
          onPrimaryAction={handlePrimaryAction}
          onRefreshWorkspace={handleRefreshWorkspace}
          isRefreshing={loadState === "loading"}
          flushContent={tplRoute.kind === 'author'}
          editMode={tplRoute.kind === 'author'}
          onLogout={handleLogout}
        >
          {workspaceView}
        </WorkspaceShell>
      )}
    </div>
  );
}

function formatDate(value?: string): string {
  if (!value) return "-";
  return new Intl.DateTimeFormat("pt-BR", { dateStyle: "short", timeStyle: "short" }).format(new Date(value));
}

function roleLabelFromRoles(roles: UserRole[]): string {
  if (roles.includes("admin") || roles.includes("system_admin")) return "Administrador";
  if (roles.includes("approver") || roles.includes("reviewer")) return "Aprovador";
  if (roles.includes("author")) return "Autor";
  if (roles.includes("editor")) return "Editor";
  return "Visualizador";
}
