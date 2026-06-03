interface InviteUserDialogProps {
  open?: boolean;
  onClose?: () => void;
}

export default function InviteUserDialog(_props: InviteUserDialogProps) {
  return <div data-testid="invite-user-dialog" />;
}
