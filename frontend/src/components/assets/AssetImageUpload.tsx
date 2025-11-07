import React, { useRef, useState } from 'react';
import {
  Box,
  Button,
  Image,
  Input,
  Text,
  VStack,
  Alert,
  AlertIcon,
  Progress,
  IconButton,
  Tooltip,
  useToast,
} from '@chakra-ui/react';
import { Upload, X, Camera } from 'lucide-react';
import { assetService } from '../../services/assetService';
import { Asset } from '../../types/asset';
import { getAssetImageUrl } from '../../utils/imageUrl';

interface AssetImageUploadProps {
  asset: Asset;
  onImageUpload?: (updatedAsset: Asset) => void;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
}

const AssetImageUpload: React.FC<AssetImageUploadProps> = ({
  asset,
  onImageUpload,
  size = 'md',
  showLabel = true,
}) => {
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const toast = useToast();

  const imageSizes = {
    sm: { width: '100px', height: '100px' },
    md: { width: '150px', height: '150px' },
    lg: { width: '200px', height: '200px' },
  };

  const validateFile = (file: File): string | null => {
    // Check file type
    const allowedTypes = ['image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      return 'File type not supported. Please upload JPEG, PNG, GIF, or WebP images only.';
    }

    // Check file size (5MB limit)
    const maxSize = 5 * 1024 * 1024; // 5MB in bytes
    if (file.size > maxSize) {
      return 'File size is too large. Maximum size is 5MB.';
    }

    return null;
  };

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const validationError = validateFile(file);
    if (validationError) {
      setError(validationError);
      return;
    }

    setError(null);
    handleUpload(file);
  };

  const handleUpload = async (file: File) => {
    if (!asset.id) {
      setError('Asset ID is required');
      return;
    }

    setUploading(true);
    setUploadProgress(0);

    try {
      // Simulate progress for better UX
      const progressInterval = setInterval(() => {
        setUploadProgress((prev) => Math.min(prev + 10, 90));
      }, 100);

      const result = await assetService.uploadAssetImage(asset.id, file);

      clearInterval(progressInterval);
      setUploadProgress(100);

      toast({
        title: 'Image uploaded successfully',
        description: 'Asset image has been updated.',
        status: 'success',
        duration: 3000,
        isClosable: true,
      });

      // Call callback with updated asset
      if (onImageUpload && result.asset) {
        onImageUpload(result.asset);
      }

      // Reset progress after a short delay
      setTimeout(() => {
        setUploadProgress(0);
      }, 1000);

    } catch (error: any) {
      setError(error.response?.data?.error || 'Failed to upload image');
      toast({
        title: 'Upload failed',
        description: error.response?.data?.error || 'Failed to upload image',
        status: 'error',
        duration: 5000,
        isClosable: true,
      });
    } finally {
      setUploading(false);
      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  const handleDrop = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    const file = event.dataTransfer.files?.[0];
    if (!file) return;

    const validationError = validateFile(file);
    if (validationError) {
      setError(validationError);
      return;
    }

    setError(null);
    handleUpload(file);
  };

  const handleDragOver = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault();
  };

  const handleButtonClick = () => {
    fileInputRef.current?.click();
  };

  const clearError = () => {
    setError(null);
  };


  return (
    <VStack spacing={4} align="stretch">
      {showLabel && (
        <Text fontSize="sm" fontWeight="medium" color="gray.700">
          Asset Image
        </Text>
      )}

      <Box position="relative">
        {/* Image Preview */}
        <Box
          width={imageSizes[size].width}
          height={imageSizes[size].height}
          border="2px dashed"
          borderColor={asset.image_path ? "gray.200" : "gray.300"}
          borderRadius="lg"
          overflow="hidden"
          position="relative"
          onDrop={handleDrop}
          onDragOver={handleDragOver}
          cursor={uploading ? "not-allowed" : "pointer"}
          onClick={!uploading ? handleButtonClick : undefined}
          bg={asset.image_path ? "white" : "gray.50"}
          display="flex"
          alignItems="center"
          justifyContent="center"
          _hover={{
            borderColor: asset.image_path ? "gray.300" : "blue.400",
            bg: asset.image_path ? "gray.50" : "gray.100",
          }}
        >
          {asset.image_path ? (
            <Image
              src={getAssetImageUrl(asset.image_path) || ''}
              alt={`${asset.name} image`}
              objectFit="cover"
              width="100%"
              height="100%"
              fallback={
                <Box
                  width="100%"
                  height="100%"
                  display="flex"
                  alignItems="center"
                  justifyContent="center"
                  bg="gray.100"
                >
                  <Camera size={32} color="gray.400" />
                </Box>
              }
            />
          ) : (
            <VStack spacing={2}>
              <Upload size={size === 'sm' ? 20 : size === 'md' ? 24 : 32} color="gray.400" />
              <Text fontSize={size === 'sm' ? 'xs' : 'sm'} color="gray.500" textAlign="center">
                Drop image here or click to upload
              </Text>
              {size !== 'sm' && (
                <Text fontSize="xs" color="gray.400" textAlign="center">
                  Max 5MB • JPEG, PNG, GIF, WebP
                </Text>
              )}
            </VStack>
          )}

          {/* Upload Progress */}
          {uploading && (
            <Box
              position="absolute"
              top="0"
              left="0"
              right="0"
              bottom="0"
              bg="blackAlpha.600"
              display="flex"
              flexDirection="column"
              alignItems="center"
              justifyContent="center"
            >
              <Progress
                value={uploadProgress}
                size="sm"
                width="80%"
                colorScheme="blue"
                mb={2}
              />
              <Text color="white" fontSize="sm">
                Uploading... {uploadProgress}%
              </Text>
            </Box>
          )}
        </Box>

        {/* Actions */}
        <Box position="absolute" top="2" right="2" display="flex" gap="1">
          {asset.image_path && !uploading && (
            <Tooltip label="Upload new image">
              <IconButton
                aria-label="Upload new image"
                icon={<Upload size={16} />}
                size="sm"
                variant="solid"
                colorScheme="blue"
                onClick={(e) => {
                  e.stopPropagation();
                  handleButtonClick();
                }}
              />
            </Tooltip>
          )}

          {error && (
            <Tooltip label="Clear error">
              <IconButton
                aria-label="Clear error"
                icon={<X size={16} />}
                size="sm"
                variant="solid"
                colorScheme="red"
                onClick={(e) => {
                  e.stopPropagation();
                  clearError();
                }}
              />
            </Tooltip>
          )}
        </Box>
      </Box>

      {/* Upload Button (visible when no image) */}
      {!asset.image_path && !uploading && (
        <Button
          leftIcon={<Upload size={16} />}
          size="sm"
          variant="outline"
          onClick={handleButtonClick}
        >
          Upload Image
        </Button>
      )}

      {/* Hidden File Input */}
      <Input
        ref={fileInputRef}
        type="file"
        accept="image/jpeg,image/jpg,image/png,image/gif,image/webp"
        onChange={handleFileSelect}
        display="none"
      />

      {/* Error Message */}
      {error && (
        <Alert status="error" size="sm" borderRadius="md">
          <AlertIcon />
          <Text fontSize="sm">{error}</Text>
        </Alert>
      )}

      {/* File Info */}
      {size !== 'sm' && (
        <Text fontSize="xs" color="gray.500" textAlign="center">
          Supported formats: JPEG, PNG, GIF, WebP • Maximum size: 5MB
        </Text>
      )}
    </VStack>
  );
};

export default AssetImageUpload;
