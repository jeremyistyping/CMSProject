'use client';

import React, { useRef } from 'react';
import {
  Box,
  Button,
  FormControl,
  FormLabel,
  HStack,
  VStack,
  Text,
  IconButton,
  Image,
  SimpleGrid,
  useColorModeValue,
  Icon,
  Center,
} from '@chakra-ui/react';
import { FiUpload, FiX, FiImage } from 'react-icons/fi';

interface PhotoPreview {
  file: File;
  preview: string;
}

interface PhotoUploadProps {
  photos: PhotoPreview[];
  onPhotosChange: (photos: PhotoPreview[]) => void;
  isDisabled?: boolean;
  maxPhotos?: number;
}

const PhotoUpload: React.FC<PhotoUploadProps> = ({
  photos,
  onPhotosChange,
  isDisabled = false,
  maxPhotos = 10,
}) => {
  const fileInputRef = useRef<HTMLInputElement>(null);
  
  const textColor = useColorModeValue('gray.800', 'var(--text-primary)');
  const borderColor = useColorModeValue('gray.200', 'var(--border-color)');
  const bgColor = useColorModeValue('gray.50', 'gray.700');
  const hoverBg = useColorModeValue('gray.100', 'gray.600');

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = event.target.files;
    if (!files) return;

    const newPhotos: PhotoPreview[] = [];
    const remainingSlots = maxPhotos - photos.length;

    // Convert FileList to array and limit to remaining slots
    Array.from(files).slice(0, remainingSlots).forEach((file) => {
      // Validate file type
      if (!file.type.startsWith('image/')) {
        return;
      }

      // Validate file size (max 10MB)
      if (file.size > 10 * 1024 * 1024) {
        return;
      }

      newPhotos.push({
        file,
        preview: URL.createObjectURL(file),
      });
    });

    onPhotosChange([...photos, ...newPhotos]);
    
    // Reset file input
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleRemovePhoto = (index: number) => {
    const newPhotos = [...photos];
    // Revoke object URL to free memory
    URL.revokeObjectURL(newPhotos[index].preview);
    newPhotos.splice(index, 1);
    onPhotosChange(newPhotos);
  };

  const handleUploadClick = () => {
    fileInputRef.current?.click();
  };

  return (
    <FormControl>
      <FormLabel color={textColor} fontSize="sm" fontWeight="semibold" mb={2}>
        Photos (Optional)
      </FormLabel>

      {/* Upload Button */}
      {photos.length < maxPhotos && (
        <Button
          leftIcon={<FiUpload />}
          onClick={handleUploadClick}
          isDisabled={isDisabled}
          variant="outline"
          borderColor={borderColor}
          size="sm"
          mb={3}
          _hover={{
            bg: hoverBg,
          }}
        >
          Upload Photos ({photos.length}/{maxPhotos})
        </Button>
      )}

      {/* Hidden File Input */}
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        multiple
        onChange={handleFileSelect}
        style={{ display: 'none' }}
      />

      {/* Photo Previews */}
      {photos.length > 0 ? (
        <SimpleGrid columns={{ base: 2, md: 3, lg: 4 }} spacing={3}>
          {photos.map((photo, index) => (
            <Box
              key={index}
              position="relative"
              borderWidth="1px"
              borderColor={borderColor}
              borderRadius="md"
              overflow="hidden"
              bg={bgColor}
              _hover={{
                shadow: 'md',
              }}
            >
              <Image
                src={photo.preview}
                alt={`Preview ${index + 1}`}
                objectFit="cover"
                w="full"
                h="120px"
              />
              <IconButton
                icon={<FiX />}
                size="xs"
                colorScheme="red"
                position="absolute"
                top={1}
                right={1}
                onClick={() => handleRemovePhoto(index)}
                aria-label="Remove photo"
                isDisabled={isDisabled}
              />
              <Box
                position="absolute"
                bottom={0}
                left={0}
                right={0}
                bg="blackAlpha.600"
                p={1}
              >
                <Text fontSize="xs" color="white" noOfLines={1} px={1}>
                  {photo.file.name}
                </Text>
              </Box>
            </Box>
          ))}
        </SimpleGrid>
      ) : (
        <Center
          borderWidth="2px"
          borderStyle="dashed"
          borderColor={borderColor}
          borderRadius="md"
          p={8}
          bg={bgColor}
        >
          <VStack spacing={2}>
            <Icon as={FiImage} boxSize={10} color="gray.400" />
            <Text fontSize="sm" color="gray.500">
              No photos uploaded
            </Text>
          </VStack>
        </Center>
      )}

      <Text fontSize="xs" color="gray.500" mt={2}>
        Max {maxPhotos} photos, 10MB each. Supported: JPG, PNG, GIF, WebP
      </Text>
    </FormControl>
  );
};

export default PhotoUpload;

