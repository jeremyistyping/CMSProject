'use client';

import React from 'react';
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  Button,
  Text,
  VStack,
  Alert,
  AlertIcon,
  AlertTitle,
  AlertDescription,
  useColorModeValue,
} from '@chakra-ui/react';
import { FiClock, FiLogIn } from 'react-icons/fi';

interface AuthExpiredModalProps {
  isOpen: boolean;
  onLoginRedirect: () => void;
}

const AuthExpiredModal: React.FC<AuthExpiredModalProps> = ({
  isOpen,
  onLoginRedirect,
}) => {
  const bgColor = useColorModeValue('white', 'gray.800');
  const borderColor = useColorModeValue('gray.200', 'gray.600');
  const [isRedirecting, setIsRedirecting] = React.useState(false);
  
  // Debug effect to track modal state
  React.useEffect(() => {
    console.log('AuthExpiredModal - isOpen changed:', isOpen);
  }, [isOpen]);
  
  // Auto-cleanup localStorage when modal opens
  React.useEffect(() => {
    if (isOpen && typeof window !== 'undefined') {
      // Clear any potentially corrupted auth data
      window.localStorage.removeItem('token');
      window.localStorage.removeItem('refreshToken');
      window.localStorage.removeItem('user');
      // Set logout timestamp to prevent auto-login
      window.localStorage.setItem('lastLogoutTime', Date.now().toString());
      console.log('AuthExpiredModal - Cleared localStorage and set logout timestamp');
    }
  }, [isOpen]);
  
  const handleLoginRedirect = () => {
    if (isRedirecting) return; // Prevent double-clicks
    
    setIsRedirecting(true);
    console.log('AuthExpiredModal - Redirecting to login');
    
    // Small delay to show loading state
    setTimeout(() => {
      onLoginRedirect();
    }, 500);
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={() => {}} // Prevent closing without action
      closeOnOverlayClick={false}
      closeOnEsc={false}
      size="md"
      isCentered
      zIndex={9999} // Higher than other modals
    >
      <ModalOverlay bg="blackAlpha.600" backdropFilter="blur(5px)" />
      <ModalContent
        bg={bgColor}
        borderRadius="xl"
        boxShadow="xl"
        border="1px solid"
        borderColor={borderColor}
        mx={4}
      >
        <ModalHeader
          bg="orange.500"
          color="white"
          borderTopRadius="xl"
          textAlign="center"
          py={4}
        >
          <VStack spacing={2}>
            <FiClock size={24} />
            <Text fontSize="lg" fontWeight="bold">
              Session Expired
            </Text>
          </VStack>
        </ModalHeader>

        <ModalBody py={6}>
          <VStack spacing={4} align="stretch">
            <Alert
              status="warning"
              variant="left-accent"
              borderRadius="md"
              bg="orange.50"
            >
              <AlertIcon />
              <VStack align="start" spacing={1} flex={1}>
                <AlertTitle fontSize="sm">Authentication Required</AlertTitle>
                <AlertDescription fontSize="sm">
                  Your session has expired for security reasons. Please login again to continue.
                </AlertDescription>
              </VStack>
            </Alert>

            <Text fontSize="sm" color="gray.600" textAlign="center">
              This helps protect your account and sensitive financial data.
            </Text>
          </VStack>
        </ModalBody>

        <ModalFooter justifyContent="center" pb={6}>
          <Button
            colorScheme="orange"
            leftIcon={isRedirecting ? undefined : <FiLogIn />}
            onClick={handleLoginRedirect}
            size="lg"
            minW="200px"
            borderRadius="lg"
            isLoading={isRedirecting}
            loadingText="Redirecting..."
            disabled={isRedirecting}
          >
            {isRedirecting ? 'Redirecting...' : 'Login Again'}
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
};

export default AuthExpiredModal;
