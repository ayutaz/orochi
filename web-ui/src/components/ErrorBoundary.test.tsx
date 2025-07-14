import { describe, it, expect, vi, beforeAll, afterAll } from 'vitest';
import { render, screen } from '../test/test-utils';
import ErrorBoundary from './ErrorBoundary';

// Component that throws an error
const ThrowError = ({ shouldThrow }: { shouldThrow: boolean }) => {
  if (shouldThrow) {
    throw new Error('Test error');
  }
  return <div>No error</div>;
};

describe('ErrorBoundary', () => {
  // Mock console.error to avoid noise in test output
  const originalError = console.error;
  beforeAll(() => {
    console.error = vi.fn();
  });

  afterAll(() => {
    console.error = originalError;
  });

  it('should render children when there is no error', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={false} />
      </ErrorBoundary>
    );

    expect(screen.getByText('No error')).toBeInTheDocument();
  });

  it('should render error UI when there is an error', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.getByText('Test error')).toBeInTheDocument();
  });

  it('should reset error state when reload button is clicked', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('Something went wrong')).toBeInTheDocument();

    // Click Go to Home button
    const goToHomeButton = screen.getByText('Go to Home');

    // Mock window.location.href setter
    const mockHrefSetter = vi.fn();
    Object.defineProperty(window, 'location', {
      value: {
        get href() {
          return 'http://localhost:3000/test';
        },
        set href(value) {
          mockHrefSetter(value);
        },
      },
      configurable: true,
    });

    goToHomeButton.click();

    // Check that we navigated to home
    expect(mockHrefSetter).toHaveBeenCalledWith('/');
  });
});
