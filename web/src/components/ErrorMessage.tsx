export function ErrorMessage({ error }: { error: Error }) {
  return (
    <p role="alert" className="text-sm text-danger">
      {error.message}
    </p>
  );
}
